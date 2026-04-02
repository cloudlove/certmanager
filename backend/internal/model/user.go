package model

import (
	"time"
)

// User 用户模型
type User struct {
	BaseModel
	Username    string     `json:"username" gorm:"uniqueIndex;size:64;not null"`
	Password    string     `json:"-" gorm:"size:128;not null"`
	Email       string     `json:"email" gorm:"size:128"`
	Nickname    string     `json:"nickname" gorm:"size:64"`
	RoleID      uint       `json:"role_id"`
	Role        Role       `json:"role" gorm:"foreignKey:RoleID"`
	Status      string     `json:"status" gorm:"size:16;default:'active'"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
