package model

import (
	"time"
)

// Domain 域名模型
type Domain struct {
	BaseModel
	Name          string     `gorm:"size:255;not null;uniqueIndex" json:"name"`
	CertificateID uint       `gorm:"index" json:"certificate_id"`
	VerifyStatus  string     `gorm:"size:50;not null" json:"verify_status"`
	LastCheckAt   *time.Time `json:"last_check_at"`
}

// TableName 指定表名
func (Domain) TableName() string {
	return "domains"
}
