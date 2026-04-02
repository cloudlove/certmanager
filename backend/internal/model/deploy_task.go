package model

// DeployTask 部署任务模型
type DeployTask struct {
	BaseModel
	Name          string `gorm:"size:255;not null" json:"name"`
	Type          string `gorm:"size:100;not null" json:"type"`
	CertificateID uint   `gorm:"index" json:"certificate_id"`
	Status        string `gorm:"size:50;not null" json:"status"`
	TotalItems    int    `json:"total_items"`
	SuccessItems  int    `json:"success_items"`
	FailedItems   int    `json:"failed_items"`
}

// TableName 指定表名
func (DeployTask) TableName() string {
	return "deploy_tasks"
}
