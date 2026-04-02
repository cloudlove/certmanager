package repository

import (
	"time"

	"certmanager-backend/internal/model"

	"gorm.io/gorm"
)

// NginxRepository Nginx 数据访问层
type NginxRepository struct {
	db *gorm.DB
}

// NewNginxRepository 创建 NginxRepository 实例
func NewNginxRepository(db *gorm.DB) *NginxRepository {
	return &NginxRepository{db: db}
}

// ==================== 集群操作 ====================

// CreateCluster 创建集群
func (r *NginxRepository) CreateCluster(cluster *model.NginxCluster) error {
	return r.db.Create(cluster).Error
}

// GetClusterByID 根据 ID 获取集群
func (r *NginxRepository) GetClusterByID(id uint) (*model.NginxCluster, error) {
	var cluster model.NginxCluster
	if err := r.db.First(&cluster, id).Error; err != nil {
		return nil, err
	}
	return &cluster, nil
}

// ListClusters 分页查询集群列表
func (r *NginxRepository) ListClusters(page, pageSize int) ([]model.NginxCluster, int64, error) {
	var clusters []model.NginxCluster
	var total int64

	query := r.db.Model(&model.NginxCluster{})

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&clusters).Error; err != nil {
		return nil, 0, err
	}

	return clusters, total, nil
}

// UpdateCluster 更新集群
func (r *NginxRepository) UpdateCluster(cluster *model.NginxCluster) error {
	return r.db.Save(cluster).Error
}

// DeleteCluster 删除集群（级联删除节点）
func (r *NginxRepository) DeleteCluster(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先删除集群下的所有节点
		if err := tx.Where("cluster_id = ?", id).Delete(&model.NginxNode{}).Error; err != nil {
			return err
		}
		// 再删除集群
		return tx.Delete(&model.NginxCluster{}, id).Error
	})
}

// ==================== 节点操作 ====================

// AddNode 添加节点
func (r *NginxRepository) AddNode(node *model.NginxNode) error {
	return r.db.Create(node).Error
}

// RemoveNode 删除节点
func (r *NginxRepository) RemoveNode(id uint) error {
	return r.db.Delete(&model.NginxNode{}, id).Error
}

// GetNodeByID 根据 ID 获取节点
func (r *NginxRepository) GetNodeByID(id uint) (*model.NginxNode, error) {
	var node model.NginxNode
	if err := r.db.First(&node, id).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

// GetNodesByClusterID 获取集群下的所有节点
func (r *NginxRepository) GetNodesByClusterID(clusterID uint) ([]model.NginxNode, error) {
	var nodes []model.NginxNode
	if err := r.db.Where("cluster_id = ?", clusterID).Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

// UpdateNodeHeartbeat 更新节点心跳和状态
func (r *NginxRepository) UpdateNodeHeartbeat(nodeID uint) error {
	now := time.Now()
	return r.db.Model(&model.NginxNode{}).Where("id = ?", nodeID).Updates(map[string]interface{}{
		"last_heartbeat": now,
		"status":         "online",
	}).Error
}

// UpdateNodeStatus 更新节点状态
func (r *NginxRepository) UpdateNodeStatus(nodeID uint, status string) error {
	return r.db.Model(&model.NginxNode{}).Where("id = ?", nodeID).Update("status", status).Error
}

// GetNodeByIPAndPort 根据 IP 和端口获取节点
func (r *NginxRepository) GetNodeByIPAndPort(ip, port string) (*model.NginxNode, error) {
	var node model.NginxNode
	if err := r.db.Where("ip = ? AND port = ?", ip, port).First(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}
