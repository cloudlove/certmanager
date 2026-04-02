package model

// AuditLog 审计日志模型
type AuditLog struct {
	BaseModel
	UserID       string `json:"userId"`
	Action       string `json:"action"`       // create, update, delete, deploy, rollback
	ResourceType string `json:"resourceType"` // certificate, credential, csr, domain, deploy_task, nginx_cluster, notification_rule
	ResourceID   uint   `json:"resourceId"`
	Detail       string `json:"detail" gorm:"type:text"`
	IP           string `json:"ip"`
}
