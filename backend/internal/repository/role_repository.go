package repository

import (
	"certmanager-backend/internal/model"

	"gorm.io/gorm"
)

// RoleRepository 角色数据访问层
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 创建 RoleRepository 实例
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create 创建角色
func (r *RoleRepository) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

// Update 更新角色
func (r *RoleRepository) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

// Delete 删除角色
func (r *RoleRepository) Delete(id uint) error {
	return r.db.Delete(&model.Role{}, id).Error
}

// GetByID 根据 ID 获取角色
func (r *RoleRepository) GetByID(id uint) (*model.Role, error) {
	var role model.Role
	if err := r.db.Preload("Permissions").First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// GetByName 根据名称获取角色
func (r *RoleRepository) GetByName(name string) (*model.Role, error) {
	var role model.Role
	if err := r.db.Preload("Permissions").Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// List 分页查询角色列表
func (r *RoleRepository) List(page, pageSize int, name string) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	query := r.db.Model(&model.Role{}).Preload("Permissions")
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// ListAll 获取所有角色（不分页）
func (r *RoleRepository) ListAll() ([]model.Role, error) {
	var roles []model.Role
	if err := r.db.Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// AssignPermissions 分配权限给角色
func (r *RoleRepository) AssignPermissions(roleID uint, permissionIDs []uint) error {
	// 首先删除角色的所有权限
	if err := r.db.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID).Error; err != nil {
		return err
	}

	// 然后添加新的权限
	if len(permissionIDs) > 0 {
		permissions := make([]model.Permission, len(permissionIDs))
		for i, id := range permissionIDs {
			permissions[i] = model.Permission{}
			permissions[i].ID = id
		}
		return r.db.Model(&model.Role{BaseModel: model.BaseModel{ID: roleID}}).Association("Permissions").Replace(permissions)
	}

	return nil
}

// ExistsByName 检查角色名是否存在
func (r *RoleRepository) ExistsByName(name string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Role{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetRolePermissions 获取角色的所有权限
func (r *RoleRepository) GetRolePermissions(roleID uint) ([]model.Permission, error) {
	var role model.Role
	if err := r.db.Preload("Permissions").First(&role, roleID).Error; err != nil {
		return nil, err
	}
	return role.Permissions, nil
}
