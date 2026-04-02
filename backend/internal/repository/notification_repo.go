package repository

import (
	"certmanager-backend/internal/model"

	"gorm.io/gorm"
)

// NotificationRepository 通知数据访问层
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 创建 NotificationRepository 实例
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// CreateRule 创建通知规则
func (r *NotificationRepository) CreateRule(rule *model.NotificationRule) error {
	return r.db.Create(rule).Error
}

// UpdateRule 更新通知规则
func (r *NotificationRepository) UpdateRule(rule *model.NotificationRule) error {
	return r.db.Save(rule).Error
}

// DeleteRule 删除通知规则
func (r *NotificationRepository) DeleteRule(id uint) error {
	return r.db.Delete(&model.NotificationRule{}, id).Error
}

// GetRuleByID 根据 ID 获取通知规则
func (r *NotificationRepository) GetRuleByID(id uint) (*model.NotificationRule, error) {
	var rule model.NotificationRule
	if err := r.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// ListRules 分页查询通知规则列表
func (r *NotificationRepository) ListRules(page, pageSize int) ([]model.NotificationRule, int64, error) {
	var rules []model.NotificationRule
	var total int64

	query := r.db.Model(&model.NotificationRule{})

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&rules).Error; err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

// ToggleRule 切换规则启用状态
func (r *NotificationRepository) ToggleRule(id uint, enabled bool) error {
	return r.db.Model(&model.NotificationRule{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// GetRulesByEventType 根据事件类型获取启用的通知规则
func (r *NotificationRepository) GetRulesByEventType(eventType string) ([]model.NotificationRule, error) {
	var rules []model.NotificationRule
	if err := r.db.Where("event_type = ? AND enabled = ?", eventType, true).Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// GetEnabledRules 获取所有启用的通知规则
func (r *NotificationRepository) GetEnabledRules() ([]model.NotificationRule, error) {
	var rules []model.NotificationRule
	if err := r.db.Where("enabled = ?", true).Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// CreateLog 创建通知日志
func (r *NotificationRepository) CreateLog(log *model.NotificationLog) error {
	return r.db.Create(log).Error
}

// ListLogs 分页查询通知日志
func (r *NotificationRepository) ListLogs(page, pageSize int, eventType string) ([]model.NotificationLog, int64, error) {
	var logs []model.NotificationLog
	var total int64

	query := r.db.Model(&model.NotificationLog{})

	// 事件类型筛选
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetLogByID 根据 ID 获取通知日志
func (r *NotificationRepository) GetLogByID(id uint) (*model.NotificationLog, error) {
	var log model.NotificationLog
	if err := r.db.First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}
