package model

import (
	"time"
)

// NotificationLog 通知日志模型
type NotificationLog struct {
	BaseModel
	RuleID    uint       `gorm:"not null;index" json:"rule_id"`
	EventType string     `gorm:"size:100;not null" json:"event_type"`
	Content   string     `gorm:"type:text" json:"content"`
	Status    string     `gorm:"size:50;not null" json:"status"`
	SentAt    *time.Time `json:"sent_at"`
}

// TableName 指定表名
func (NotificationLog) TableName() string {
	return "notification_logs"
}
