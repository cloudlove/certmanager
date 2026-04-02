package model

// DeployTaskItem 部署任务项模型
type DeployTaskItem struct {
	BaseModel
	DeployTaskID uint   `gorm:"not null;index" json:"deploy_task_id"`
	TargetType   string `gorm:"size:100;not null" json:"target_type"`
	ProviderType string `gorm:"size:100" json:"provider_type"`
	ResourceID   string `gorm:"size:255" json:"resource_id"`
	CredentialID uint   `gorm:"index" json:"credential_id"`
	Status       string `gorm:"size:50;not null" json:"status"`
	ErrorMessage string `gorm:"type:text" json:"error_message"`
	SnapshotID   uint   `json:"snapshot_id"`
}

// TableName 指定表名
func (DeployTaskItem) TableName() string {
	return "deploy_task_items"
}
