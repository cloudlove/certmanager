package model

import (
	"time"
)

// NginxNode Nginx 节点模型
type NginxNode struct {
	BaseModel
	ClusterID     uint       `gorm:"not null;index" json:"cluster_id"`
	IP            string     `gorm:"size:50;not null" json:"ip"`
	Port          string     `gorm:"size:10;not null" json:"port"`
	Status        string     `gorm:"size:50;not null" json:"status"`
	LastHeartbeat *time.Time `json:"last_heartbeat"`
}

// TableName 指定表名
func (NginxNode) TableName() string {
	return "nginx_nodes"
}
