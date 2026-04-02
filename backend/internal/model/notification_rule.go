package model

// NotificationRule 通知规则模型
type NotificationRule struct {
	BaseModel
	Name          string `gorm:"size:255;not null" json:"name"`
	EventType     string `gorm:"size:100;not null" json:"event_type"`
	ThresholdDays int    `json:"threshold_days"`
	Channels      string `gorm:"size:500" json:"channels"`    // 逗号分隔的渠道列表
	Recipients    string `gorm:"size:1000" json:"recipients"` // 逗号分隔的接收人列表
	Enabled       bool   `gorm:"default:true" json:"enabled"`
}

// TableName 指定表名
func (NotificationRule) TableName() string {
	return "notification_rules"
}
