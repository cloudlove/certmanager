package scheduler

import (
	"fmt"
	"log"
	"strings"
	"time"

	"certmanager-backend/internal/model"
	"certmanager-backend/internal/notify"
	"certmanager-backend/internal/service"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	cron      *cron.Cron
	db        *gorm.DB
	notifySvc *service.NotificationService
	certSvc   *service.CertificateService
}

// NewScheduler 创建 Scheduler 实例
func NewScheduler(db *gorm.DB, notifySvc *service.NotificationService, certSvc *service.CertificateService) *Scheduler {
	return &Scheduler{
		cron:      cron.New(),
		db:        db,
		notifySvc: notifySvc,
		certSvc:   certSvc,
	}
}

// Start 启动定时任务
func (s *Scheduler) Start() {
	// 每日 09:00 扫描即将到期证书
	_, err := s.cron.AddFunc("0 9 * * *", s.checkExpiryCerts)
	if err != nil {
		log.Printf("Failed to add expiry check cron job: %v", err)
	}

	// 每 5 分钟同步待处理证书状态
	_, err = s.cron.AddFunc("*/5 * * * *", s.syncPendingCerts)
	if err != nil {
		log.Printf("Failed to add cert sync cron job: %v", err)
	}

	// 启动定时任务
	s.cron.Start()
	log.Println("Scheduler started")
}

// Stop 停止定时任务
func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("Scheduler stopped")
}

// checkExpiryCerts 检查即将到期的证书
func (s *Scheduler) checkExpiryCerts() {
	log.Println("Checking expiry certificates...")

	// 获取所有启用的证书到期通知规则
	var rules []model.NotificationRule
	if err := s.db.Where("event_type = ? AND enabled = ?", "cert_expiry", true).Find(&rules).Error; err != nil {
		log.Printf("Failed to get expiry notification rules: %v", err)
		return
	}

	if len(rules) == 0 {
		log.Println("No expiry notification rules found")
		return
	}

	now := time.Now()

	for _, rule := range rules {
		// 计算阈值日期
		thresholdDate := now.AddDate(0, 0, rule.ThresholdDays)

		// 查询在阈值天数内到期的证书
		var certs []model.Certificate
		if err := s.db.Where("expire_at <= ? AND expire_at > ?", thresholdDate, now).Find(&certs).Error; err != nil {
			log.Printf("Failed to get expiring certificates: %v", err)
			continue
		}

		if len(certs) == 0 {
			continue
		}

		// 构建通知内容
		title := fmt.Sprintf("证书到期提醒 - %d 个证书即将到期", len(certs))
		content := s.buildExpiryContent(certs)

		// 发送通知
		if err := s.sendNotification(&rule, title, content); err != nil {
			log.Printf("Failed to send expiry notification: %v", err)
		}
	}

	log.Println("Expiry certificates check completed")
}

// buildExpiryContent 构建到期通知内容
func (s *Scheduler) buildExpiryContent(certs []model.Certificate) string {
	var sb strings.Builder
	sb.WriteString("以下证书即将到期，请及时处理:\n\n")

	for i, cert := range certs {
		daysUntilExpiry := int(cert.ExpireAt.Sub(time.Now()).Hours() / 24)
		sb.WriteString(fmt.Sprintf("%d. 域名: %s\n", i+1, cert.Domain))
		sb.WriteString(fmt.Sprintf("   颁发者: %s\n", cert.Issuer))
		sb.WriteString(fmt.Sprintf("   到期时间: %s\n", cert.ExpireAt.Format("2006-01-02")))
		sb.WriteString(fmt.Sprintf("   剩余天数: %d 天\n\n", daysUntilExpiry))
	}

	sb.WriteString("请登录证书管理系统查看详情并处理。")
	return sb.String()
}

// sendNotification 发送通知
func (s *Scheduler) sendNotification(rule *model.NotificationRule, title, content string) error {
	channels := strings.Split(rule.Channels, ",")
	recipients := strings.Split(rule.Recipients, ",")

	msg := notify.NotifyMessage{
		Title:      title,
		Content:    content,
		Recipients: recipients,
	}

	for _, channel := range channels {
		channel = strings.TrimSpace(channel)
		if channel == "" {
			continue
		}

		var config map[string]string
		if channel == "webhook" || channel == "dingtalk" || channel == "wecom" {
			config = map[string]string{
				"webhook_url": recipients[0],
			}
		}

		notifier, err := notify.NewNotifier(channel, config)
		if err != nil {
			s.createLog(rule.ID, rule.EventType, fmt.Sprintf("[%s] %s", channel, content), "failed")
			continue
		}

		if err := notifier.Send(nil, msg); err != nil {
			s.createLog(rule.ID, rule.EventType, fmt.Sprintf("[%s] %s", channel, content), "failed")
			log.Printf("Failed to send %s notification: %v", channel, err)
		} else {
			s.createLog(rule.ID, rule.EventType, fmt.Sprintf("[%s] %s", channel, content), "success")
			log.Printf("%s notification sent successfully", channel)
		}
	}

	return nil
}

// createLog 创建通知日志
func (s *Scheduler) createLog(ruleID uint, eventType, content, status string) {
	now := time.Now()
	notificationLog := &model.NotificationLog{
		RuleID:    ruleID,
		EventType: eventType,
		Content:   content,
		Status:    status,
		SentAt:    &now,
	}
	if err := s.db.Create(notificationLog).Error; err != nil {
		log.Printf("Failed to create notification log: %v", err)
	}
}

// syncPendingCerts 同步待处理证书状态
func (s *Scheduler) syncPendingCerts() {
	log.Println("Starting pending certificates sync...")
	if s.certSvc != nil {
		s.certSvc.SyncPendingCertificates()
	}
}
