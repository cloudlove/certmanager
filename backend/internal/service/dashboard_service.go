package service

import (
	"fmt"
	"time"

	"certmanager-backend/internal/model"

	"gorm.io/gorm"
)

// DashboardService 大盘统计服务
type DashboardService struct {
	db *gorm.DB
}

// NewDashboardService 创建 DashboardService 实例
func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

// CertOverview 证书概览统计
type CertOverview struct {
	Total      int64 `json:"total"`
	Issued     int64 `json:"issued"`
	Pending    int64 `json:"pending"`
	Expired    int64 `json:"expired"`
	Revoked    int64 `json:"revoked"`
	Expiring7  int64 `json:"expiring_7"`  // 7天内到期
	Expiring15 int64 `json:"expiring_15"` // 15天内到期
	Expiring30 int64 `json:"expiring_30"` // 30天内到期
}

// DeployOverview 部署概览统计
type DeployOverview struct {
	TotalTasks30d int64        `json:"total_tasks_30d"`
	SuccessRate   float64      `json:"success_rate"`
	FailedTop5    []FailedItem `json:"failed_top_5"`
}

// FailedItem 失败项统计
type FailedItem struct {
	Domain    string `json:"domain"`
	FailCount int64  `json:"fail_count"`
}

// CloudDistItem 云资源分布项
type CloudDistItem struct {
	Provider    string `json:"provider"`
	CertCount   int64  `json:"cert_count"`
	DeployCount int64  `json:"deploy_count"`
}

// ExpiryTrendItem 到期趋势项
type ExpiryTrendItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// AlertItem 告警项
type AlertItem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Level  string `json:"level"`
}

// GetCertOverview 获取证书概览统计
func (s *DashboardService) GetCertOverview() (*CertOverview, error) {
	overview := &CertOverview{}

	// 统计总数
	if err := s.db.Model(&model.Certificate{}).Count(&overview.Total).Error; err != nil {
		return nil, err
	}

	// 按状态统计
	var statusCounts []struct {
		Status string
		Count  int64
	}
	if err := s.db.Model(&model.Certificate{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return nil, err
	}

	for _, sc := range statusCounts {
		switch sc.Status {
		case "issued":
			overview.Issued = sc.Count
		case "pending":
			overview.Pending = sc.Count
		case "expired":
			overview.Expired = sc.Count
		case "revoked":
			overview.Revoked = sc.Count
		}
	}

	now := time.Now()

	// 7天内到期
	if err := s.db.Model(&model.Certificate{}).
		Where("expire_at > ? AND expire_at <= ?", now, now.AddDate(0, 0, 7)).
		Count(&overview.Expiring7).Error; err != nil {
		return nil, err
	}

	// 15天内到期
	if err := s.db.Model(&model.Certificate{}).
		Where("expire_at > ? AND expire_at <= ?", now, now.AddDate(0, 0, 15)).
		Count(&overview.Expiring15).Error; err != nil {
		return nil, err
	}

	// 30天内到期
	if err := s.db.Model(&model.Certificate{}).
		Where("expire_at > ? AND expire_at <= ?", now, now.AddDate(0, 0, 30)).
		Count(&overview.Expiring30).Error; err != nil {
		return nil, err
	}

	return overview, nil
}

// GetDeployOverview 获取部署概览统计
func (s *DashboardService) GetDeployOverview() (*DeployOverview, error) {
	overview := &DeployOverview{
		FailedTop5: make([]FailedItem, 0),
	}

	// 近30天任务数
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	if err := s.db.Model(&model.DeployTask{}).
		Where("created_at >= ?", thirtyDaysAgo).
		Count(&overview.TotalTasks30d).Error; err != nil {
		return nil, err
	}

	// 计算成功率
	var successCount, totalCount int64
	if err := s.db.Model(&model.DeployTask{}).
		Where("created_at >= ?", thirtyDaysAgo).
		Count(&totalCount).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&model.DeployTask{}).
		Where("created_at >= ? AND status = ?", thirtyDaysAgo, "success").
		Count(&successCount).Error; err != nil {
		return nil, err
	}

	if totalCount > 0 {
		overview.SuccessRate = float64(successCount) * 100 / float64(totalCount)
	} else {
		overview.SuccessRate = 100.0
	}

	// 失败 TOP5 (按域名统计)
	var failedItems []struct {
		CertificateID uint  `gorm:"column:certificate_id"`
		FailCount     int64 `gorm:"column:fail_count"`
	}
	if err := s.db.Model(&model.DeployTask{}).
		Select("certificate_id, COUNT(*) as fail_count").
		Where("created_at >= ? AND status = ?", thirtyDaysAgo, "failed").
		Group("certificate_id").
		Order("fail_count DESC").
		Limit(5).
		Scan(&failedItems).Error; err != nil {
		return nil, err
	}

	// 获取域名信息
	for _, item := range failedItems {
		var cert model.Certificate
		if err := s.db.Select("domain").First(&cert, item.CertificateID).Error; err != nil {
			continue
		}
		overview.FailedTop5 = append(overview.FailedTop5, FailedItem{
			Domain:    cert.Domain,
			FailCount: item.FailCount,
		})
	}

	return overview, nil
}

// GetCloudDistribution 获取云资源分布
func (s *DashboardService) GetCloudDistribution() ([]CloudDistItem, error) {
	// 获取所有云提供商列表
	var providers []string
	if err := s.db.Model(&model.Certificate{}).
		Select("DISTINCT ca_provider").
		Where("ca_provider != ?", "").
		Pluck("ca_provider", &providers).Error; err != nil {
		return nil, err
	}

	result := make([]CloudDistItem, 0)

	for _, provider := range providers {
		item := CloudDistItem{Provider: provider}

		// 统计该提供商的证书数量
		if err := s.db.Model(&model.Certificate{}).
			Where("ca_provider = ?", provider).
			Count(&item.CertCount).Error; err != nil {
			continue
		}

		// 统计该提供商的部署数量
		if err := s.db.Model(&model.DeployTaskItem{}).
			Where("provider_type = ?", provider).
			Count(&item.DeployCount).Error; err != nil {
			continue
		}

		result = append(result, item)
	}

	return result, nil
}

// GetNginxClusterStatus 获取 Nginx 集群状态
func (s *DashboardService) GetNginxClusterStatus() (map[string]interface{}, error) {
	var clusterCount, nodeCount, onlineCount int64

	// 集群数
	if err := s.db.Model(&model.NginxCluster{}).Count(&clusterCount).Error; err != nil {
		return nil, err
	}

	// 总节点数
	if err := s.db.Model(&model.NginxNode{}).Count(&nodeCount).Error; err != nil {
		return nil, err
	}

	// 在线节点数
	if err := s.db.Model(&model.NginxNode{}).
		Where("status = ?", "online").
		Count(&onlineCount).Error; err != nil {
		return nil, err
	}

	onlineRate := 0.0
	if nodeCount > 0 {
		onlineRate = float64(onlineCount) * 100 / float64(nodeCount)
	}

	return map[string]interface{}{
		"cluster_count": clusterCount,
		"node_count":    nodeCount,
		"online_count":  onlineCount,
		"online_rate":   onlineRate,
	}, nil
}

// GetExpiryTrend 获取证书到期趋势
func (s *DashboardService) GetExpiryTrend(days int) ([]ExpiryTrendItem, error) {
	if days <= 0 {
		days = 90
	}

	result := make([]ExpiryTrendItem, days)
	now := time.Now()

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")

		var count int64
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.AddDate(0, 0, 1)

		if err := s.db.Model(&model.Certificate{}).
			Where("expire_at >= ? AND expire_at < ?", startOfDay, endOfDay).
			Count(&count).Error; err != nil {
			return nil, err
		}

		result[i] = ExpiryTrendItem{
			Date:  dateStr,
			Count: count,
		}
	}

	return result, nil
}

// GetRecentDeployTasks 获取最近部署任务
func (s *DashboardService) GetRecentDeployTasks(limit int) ([]model.DeployTask, error) {
	if limit <= 0 {
		limit = 10
	}

	var tasks []model.DeployTask
	if err := s.db.Order("created_at DESC").Limit(limit).Find(&tasks).Error; err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetAlerts 获取告警列表
func (s *DashboardService) GetAlerts() ([]AlertItem, error) {
	alerts := make([]AlertItem, 0)
	now := time.Now()

	// 1. 已过期证书
	var expiredCerts []model.Certificate
	if err := s.db.Where("expire_at < ?", now).Limit(10).Find(&expiredCerts).Error; err == nil {
		for _, cert := range expiredCerts {
			alerts = append(alerts, AlertItem{
				Type:   "expired_cert",
				Title:  fmt.Sprintf("证书已过期: %s", cert.Domain),
				Detail: fmt.Sprintf("证书已于 %s 过期", cert.ExpireAt.Format("2006-01-02")),
				Level:  "error",
			})
		}
	}

	// 2. 即将过期证书 (7天内)
	var expiringCerts []model.Certificate
	if err := s.db.Where("expire_at > ? AND expire_at <= ?", now, now.AddDate(0, 0, 7)).Limit(10).Find(&expiringCerts).Error; err == nil {
		for _, cert := range expiringCerts {
			daysUntilExpiry := int(cert.ExpireAt.Sub(now).Hours() / 24)
			alerts = append(alerts, AlertItem{
				Type:   "expiring_cert",
				Title:  fmt.Sprintf("证书即将过期: %s", cert.Domain),
				Detail: fmt.Sprintf("证书将在 %d 天后过期 (%s)", daysUntilExpiry, cert.ExpireAt.Format("2006-01-02")),
				Level:  "warning",
			})
		}
	}

	// 3. 校验失败的域名
	var failedDomains []model.Domain
	if err := s.db.Where("verify_status IN ?", []string{"error", "mismatch"}).Limit(10).Find(&failedDomains).Error; err == nil {
		for _, domain := range failedDomains {
			alerts = append(alerts, AlertItem{
				Type:   "verify_failed",
				Title:  fmt.Sprintf("域名校验失败: %s", domain.Name),
				Detail: "域名验证未通过，请检查 DNS 配置",
				Level:  "warning",
			})
		}
	}

	// 4. 离线节点
	var offlineNodes []model.NginxNode
	if err := s.db.Where("status != ?", "online").Limit(10).Find(&offlineNodes).Error; err == nil {
		for _, node := range offlineNodes {
			alerts = append(alerts, AlertItem{
				Type:   "node_offline",
				Title:  fmt.Sprintf("Nginx 节点离线: %s", node.IP),
				Detail: fmt.Sprintf("节点状态: %s", node.Status),
				Level:  "error",
			})
		}
	}

	return alerts, nil
}

// GetFullOverview 获取完整概览数据
func (s *DashboardService) GetFullOverview() (map[string]interface{}, error) {
	certOverview, err := s.GetCertOverview()
	if err != nil {
		return nil, err
	}

	deployOverview, err := s.GetDeployOverview()
	if err != nil {
		return nil, err
	}

	cloudDist, err := s.GetCloudDistribution()
	if err != nil {
		return nil, err
	}

	nginxStatus, err := s.GetNginxClusterStatus()
	if err != nil {
		return nil, err
	}

	expiryTrend, err := s.GetExpiryTrend(90)
	if err != nil {
		return nil, err
	}

	recentTasks, err := s.GetRecentDeployTasks(10)
	if err != nil {
		return nil, err
	}

	alerts, err := s.GetAlerts()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"cert_overview":      certOverview,
		"deploy_overview":    deployOverview,
		"cloud_distribution": cloudDist,
		"nginx_status":       nginxStatus,
		"expiry_trend":       expiryTrend,
		"recent_tasks":       recentTasks,
		"alerts":             alerts,
	}, nil
}
