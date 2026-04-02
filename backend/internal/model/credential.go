package model

// CloudCredential 云凭证模型
type CloudCredential struct {
	BaseModel
	Name               string `gorm:"size:255;not null" json:"name"`
	ProviderType       string `gorm:"size:100;not null" json:"provider_type"`
	AccessKeyEncrypted string `gorm:"type:text" json:"-"`
	SecretKeyEncrypted string `gorm:"type:text" json:"-"`
	ExtraConfig        string `gorm:"type:text" json:"extra_config"`
	Status             string `gorm:"size:50;not null" json:"status"`
}

// TableName 指定表名
func (CloudCredential) TableName() string {
	return "cloud_credentials"
}
