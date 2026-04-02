package model

// Permission 权限模型
type Permission struct {
	BaseModel
	Name     string `json:"name" gorm:"uniqueIndex;size:128;not null"`
	Resource string `json:"resource" gorm:"size:64;not null"`
	Action   string `json:"action" gorm:"size:32;not null"`
	MenuKey  string `json:"menu_key" gorm:"size:64"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}
