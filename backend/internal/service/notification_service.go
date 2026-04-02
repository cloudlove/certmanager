package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"certmanager-backend/internal/model"
	"certmanager-backend/internal/notify"
	"certmanager-backend/internal/repository"
)

// NotificationService 通知服务层
type NotificationService struct {
	repo *repository.NotificationRepository
}

// NewNotificationService 创建 NotificationService 实例
func NewNotificationService(repo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

// CreateRuleRequest 创建规则请求
type CreateRuleRequest struct {
	Name          string   `json:"name"`
	EventType     string   `json:"event_type"`
	ThresholdDays int      `json:"threshold_days"`
	Channels      []string `json:"channels"`
	Recipients    []string `json:"recipients"`
	Enabled       bool     `json:"enabled"`
}

// UpdateRuleRequest 更新规则请求
type UpdateRuleRequest struct {
	Name          string   `json:"name"`
	EventType     string   `json:"event_type"`
	ThresholdDays int      `json:"threshold_days"`
	Channels      []string `json:"channels"`
	Recipients    []string `json:"recipients"`
	Enabled       bool     `json:"enabled"`
}

// NotificationRuleVO 通知规则视图对象
type NotificationRuleVO struct {
	ID            uint     `json:"id"`
	Name          string   `json:"name"`
	EventType     string   `json:"event_type"`
	ThresholdDays int      `json:"threshold_days"`
	Channels      []string `json:"channels"`
	Recipients    []string `json:"recipients"`
	Enabled       bool     `json:"enabled"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// NotificationLogVO 通知日志视图对象
type NotificationLogVO struct {
	ID        uint   `json:"id"`
	RuleID    uint   `json:"rule_id"`
	EventType string `json:"event_type"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	SentAt    string `json:"sent_at"`
	CreatedAt string `json:"created_at"`
}

// CreateRule 创建通知规则
func (s *NotificationService) CreateRule(req *CreateRuleRequest) (*NotificationRuleVO, error) {
	rule := &model.NotificationRule{
		Name:          req.Name,
		EventType:     req.EventType,
		ThresholdDays: req.ThresholdDays,
		Channels:      strings.Join(req.Channels, ","),
		Recipients:    strings.Join(req.Recipients, ","),
		Enabled:       req.Enabled,
	}

	if err := s.repo.CreateRule(rule); err != nil {
		return nil, fmt.Errorf("failed to create notification rule: %w", err)
	}

	return s.toRuleVO(rule), nil
}

// UpdateRule 更新通知规则
func (s *NotificationService) UpdateRule(id uint, req *UpdateRuleRequest) (*NotificationRuleVO, error) {
	rule, err := s.repo.GetRuleByID(id)
	if err != nil {
		return nil, fmt.Errorf("notification rule not found: %w", err)
	}

	rule.Name = req.Name
	rule.EventType = req.EventType
	rule.ThresholdDays = req.ThresholdDays
	rule.Channels = strings.Join(req.Channels, ",")
	rule.Recipients = strings.Join(req.Recipients, ",")
	rule.Enabled = req.Enabled

	if err := s.repo.UpdateRule(rule); err != nil {
		return nil, fmt.Errorf("failed to update notification rule: %w", err)
	}

	return s.toRuleVO(rule), nil
}

// DeleteRule 删除通知规则
func (s *NotificationService) DeleteRule(id uint) error {
	return s.repo.DeleteRule(id)
}

// GetRule 获取通知规则详情
func (s *NotificationService) GetRule(id uint) (*NotificationRuleVO, error) {
	rule, err := s.repo.GetRuleByID(id)
	if err != nil {
		return nil, fmt.Errorf("notification rule not found: %w", err)
	}
	return s.toRuleVO(rule), nil
}

// ListRules 获取通知规则列表
func (s *NotificationService) ListRules(page, pageSize int) ([]*NotificationRuleVO, int64, error) {
	rules, total, err := s.repo.ListRules(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	vos := make([]*NotificationRuleVO, len(rules))
	for i, rule := range rules {
		vos[i] = s.toRuleVO(&rule)
	}

	return vos, total, nil
}

// ToggleRule 切换规则启用状态
func (s *NotificationService) ToggleRule(id uint, enabled bool) error {
	return s.repo.ToggleRule(id, enabled)
}

// TestRule 测试通知规则
func (s *NotificationService) TestRule(id uint) error {
	rule, err := s.repo.GetRuleByID(id)
	if err != nil {
		return fmt.Errorf("notification rule not found: %w", err)
	}

	channels := strings.Split(rule.Channels, ",")
	recipients := strings.Split(rule.Recipients, ",")

	msg := notify.NotifyMessage{
		Title:      "测试通知",
		Content:    fmt.Sprintf("这是一条来自证书管理系统的测试通知，规则名称: %s", rule.Name),
		Recipients: recipients,
	}

	ctx := context.Background()
	for _, channel := range channels {
		channel = strings.TrimSpace(channel)
		if channel == "" {
			continue
		}

		notifier, err := notify.NewNotifier(channel, map[string]string{
			"webhook_url": recipients[0], // 使用第一个接收人作为 webhook URL
		})
		if err != nil {
			s.createLog(rule.ID, rule.EventType, fmt.Sprintf("[%s] %s", channel, msg.Content), "failed")
			continue
		}

		if err := notifier.Send(ctx, msg); err != nil {
			s.createLog(rule.ID, rule.EventType, fmt.Sprintf("[%s] %s", channel, msg.Content), "failed")
		} else {
			s.createLog(rule.ID, rule.EventType, fmt.Sprintf("[%s] %s", channel, msg.Content), "success")
		}
	}

	return nil
}

// SendNotification 发送通知
func (s *NotificationService) SendNotification(eventType string, title, content string) error {
	rules, err := s.repo.GetRulesByEventType(eventType)
	if err != nil {
		return fmt.Errorf("failed to get notification rules: %w", err)
	}

	if len(rules) == 0 {
		return nil // 没有匹配的规则
	}

	ctx := context.Background()
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

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
				s.createLog(rule.ID, eventType, fmt.Sprintf("[%s] %s", channel, content), "failed")
				continue
			}

			if err := notifier.Send(ctx, msg); err != nil {
				s.createLog(rule.ID, eventType, fmt.Sprintf("[%s] %s", channel, content), "failed")
			} else {
				s.createLog(rule.ID, eventType, fmt.Sprintf("[%s] %s", channel, content), "success")
			}
		}
	}

	return nil
}

// ListLogs 获取通知日志列表
func (s *NotificationService) ListLogs(page, pageSize int, eventType string) ([]*NotificationLogVO, int64, error) {
	logs, total, err := s.repo.ListLogs(page, pageSize, eventType)
	if err != nil {
		return nil, 0, err
	}

	vos := make([]*NotificationLogVO, len(logs))
	for i, log := range logs {
		vos[i] = s.toLogVO(&log)
	}

	return vos, total, nil
}

// GetRulesByEventType 根据事件类型获取规则
func (s *NotificationService) GetRulesByEventType(eventType string) ([]*NotificationRuleVO, error) {
	rules, err := s.repo.GetRulesByEventType(eventType)
	if err != nil {
		return nil, err
	}

	vos := make([]*NotificationRuleVO, len(rules))
	for i, rule := range rules {
		vos[i] = s.toRuleVO(&rule)
	}

	return vos, nil
}

// createLog 创建通知日志
func (s *NotificationService) createLog(ruleID uint, eventType, content, status string) {
	now := time.Now()
	log := &model.NotificationLog{
		RuleID:    ruleID,
		EventType: eventType,
		Content:   content,
		Status:    status,
		SentAt:    &now,
	}
	_ = s.repo.CreateLog(log)
}

// toRuleVO 转换为规则视图对象
func (s *NotificationService) toRuleVO(rule *model.NotificationRule) *NotificationRuleVO {
	channels := []string{}
	if rule.Channels != "" {
		channels = strings.Split(rule.Channels, ",")
	}

	recipients := []string{}
	if rule.Recipients != "" {
		recipients = strings.Split(rule.Recipients, ",")
	}

	return &NotificationRuleVO{
		ID:            rule.ID,
		Name:          rule.Name,
		EventType:     rule.EventType,
		ThresholdDays: rule.ThresholdDays,
		Channels:      channels,
		Recipients:    recipients,
		Enabled:       rule.Enabled,
		CreatedAt:     rule.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     rule.UpdatedAt.Format(time.RFC3339),
	}
}

// toLogVO 转换为日志视图对象
func (s *NotificationService) toLogVO(log *model.NotificationLog) *NotificationLogVO {
	sentAt := ""
	if log.SentAt != nil {
		sentAt = log.SentAt.Format(time.RFC3339)
	}

	return &NotificationLogVO{
		ID:        log.ID,
		RuleID:    log.RuleID,
		EventType: log.EventType,
		Content:   log.Content,
		Status:    log.Status,
		SentAt:    sentAt,
		CreatedAt: log.CreatedAt.Format(time.RFC3339),
	}
}
