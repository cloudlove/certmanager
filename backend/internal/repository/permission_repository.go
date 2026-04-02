package repository

import (
	"certmanager-backend/internal/model"

	"gorm.io/gorm"
)

// PermissionRepository 权限数据访问层
type PermissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository 创建 PermissionRepository 实例
func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// Create 创建权限
func (r *PermissionRepository) Create(permission *model.Permission) error {
	return r.db.Create(permission).Error
}

// CreateBatch 批量创建权限
func (r *PermissionRepository) CreateBatch(permissions []model.Permission) error {
	return r.db.Create(&permissions).Error
}

// GetByID 根据 ID 获取权限
func (r *PermissionRepository) GetByID(id uint) (*model.Permission, error) {
	var permission model.Permission
	if err := r.db.First(&permission, id).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

// GetByResourceAndAction 根据 resource 和 action 获取权限
func (r *PermissionRepository) GetByResourceAndAction(resource, action string) (*model.Permission, error) {
	var permission model.Permission
	if err := r.db.Where("resource = ? AND action = ?", resource, action).First(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

// List 获取所有权限
func (r *PermissionRepository) List() ([]model.Permission, error) {
	var permissions []model.Permission
	if err := r.db.Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// ListByResource 根据资源获取权限列表
func (r *PermissionRepository) ListByResource(resource string) ([]model.Permission, error) {
	var permissions []model.Permission
	if err := r.db.Where("resource = ?", resource).Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// Delete 删除权限
func (r *PermissionRepository) Delete(id uint) error {
	return r.db.Delete(&model.Permission{}, id).Error
}

// ExistsByResourceAndAction 检查权限是否存在
func (r *PermissionRepository) ExistsByResourceAndAction(resource, action string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Permission{}).Where("resource = ? AND action = ?", resource, action).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetPermissionsByRoleID 获取角色的所有权限
func (r *PermissionRepository) GetPermissionsByRoleID(roleID uint) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}
