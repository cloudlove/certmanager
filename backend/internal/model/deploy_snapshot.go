package model

// DeploySnapshot 部署快照模型
type DeploySnapshot struct {
	BaseModel
	DeployTaskItemID uint   `gorm:"not null;index" json:"deploy_task_item_id"`
	OldCertInfo      string `gorm:"type:text" json:"old_cert_info"`
}

// TableName 指定表名
func (DeploySnapshot) TableName() string {
	return "deploy_snapshots"
}
