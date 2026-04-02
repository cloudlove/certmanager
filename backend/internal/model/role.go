package model

// Role 角色模型
type Role struct {
	BaseModel
	Name        string       `json:"name" gorm:"uniqueIndex;size:64;not null"`
	Description string       `json:"description" gorm:"size:256"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}
