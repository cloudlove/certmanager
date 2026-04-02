package model

import (
	"time"
)

// Certificate 证书模型
type Certificate struct {
	BaseModel
	Domain              string    `gorm:"size:255;not null;index" json:"domain"`
	CAProvider          string    `gorm:"size:100" json:"ca_provider"`
	Status              string    `gorm:"size:50;not null" json:"status"`
	ExpireAt            time.Time `json:"expire_at"`
	CSRID               uint      `gorm:"index" json:"csr_id"`
	CredentialID        uint      `gorm:"index" json:"credential_id"` // 关联的云凭证 ID
	CertPEM             string    `gorm:"type:text" json:"-"`
	ChainPEM            string    `gorm:"type:text" json:"-"`
	PrivateKeyEncrypted string    `gorm:"type:text" json:"-"`
	Issuer              string    `gorm:"size:255" json:"issuer"`
	Fingerprint         string    `gorm:"size:255;index" json:"fingerprint"`
	KeyAlgorithm        string    `gorm:"size:50" json:"key_algorithm"`
	SerialNumber        string    `gorm:"size:255" json:"serial_number"`
	OrderID             string    `gorm:"size:255" json:"order_id"`     // CA返回的订单ID，用于后续状态查询
	VerifyType          string    `gorm:"size:20" json:"verify_type"`   // 验证方式：DNS / FILE
	VerifyInfo          string    `gorm:"type:text" json:"verify_info"` // JSON格式的验证信息
	ProductType         string    `gorm:"size:20" json:"product_type"`  // 证书级别：DV / OV / EV
	DomainType          string    `gorm:"size:20" json:"domain_type"`   // 域名类型：single / wildcard / multi
}

// TableName 指定表名
func (Certificate) TableName() string {
	return "certificates"
}
