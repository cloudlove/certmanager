package service

import (
	"errors"
	"fmt"

	"certmanager-backend/internal/model"
	"certmanager-backend/internal/repository"

	"gorm.io/gorm"
)

// RoleService 角色服务层
type RoleService struct {
	roleRepo       *repository.RoleRepository
	permissionRepo *repository.PermissionRepository
}

// NewRoleService 创建 RoleService 实例
func NewRoleService(roleRepo *repository.RoleRepository, permissionRepo *repository.PermissionRepository) *RoleService {
	return &RoleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// PermissionVO 权限视图对象
type PermissionVO struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
	MenuKey  string `json:"menu_key"`
}

// RoleVO 角色视图对象
type RoleVO struct {
	ID          uint           `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Permissions []PermissionVO `json:"permissions"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	PermissionIDs []uint `json:"permission_ids"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	PermissionIDs []uint `json:"permission_ids"`
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(req *CreateRoleRequest) (*RoleVO, error) {
	// 检查角色名是否存在
	exists, err := s.roleRepo.ExistsByName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("检查角色名失败: %w", err)
	}
	if exists {
		return nil, errors.New("角色名已存在")
	}

	role := &model.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, fmt.Errorf("创建角色失败: %w", err)
	}

	// 分配权限
	if len(req.PermissionIDs) > 0 {
		if err := s.roleRepo.AssignPermissions(role.ID, req.PermissionIDs); err != nil {
			return nil, fmt.Errorf("分配权限失败: %w", err)
		}
	}

	// 重新获取角色
	role, err = s.roleRepo.GetByID(role.ID)
	if err != nil {
		return nil, fmt.Errorf("获取角色失败: %w", err)
	}

	return s.toRoleVO(role), nil
}

// UpdateRole 更新角色
func (s *RoleService) UpdateRole(id uint, req *UpdateRoleRequest) (*RoleVO, error) {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("角色不存在: %w", err)
	}

	// 检查角色名是否已被其他角色使用
	if req.Name != "" && req.Name != role.Name {
		exists, err := s.roleRepo.ExistsByName(req.Name)
		if err != nil {
			return nil, fmt.Errorf("检查角色名失败: %w", err)
		}
		if exists {
			return nil, errors.New("角色名已存在")
		}
		role.Name = req.Name
	}

	if req.Description != "" {
		role.Description = req.Description
	}

	if err := s.roleRepo.Update(role); err != nil {
		return nil, fmt.Errorf("更新角色失败: %w", err)
	}

	// 更新权限
	if req.PermissionIDs != nil {
		if err := s.roleRepo.AssignPermissions(id, req.PermissionIDs); err != nil {
			return nil, fmt.Errorf("分配权限失败: %w", err)
		}
	}

	// 重新获取角色
	role, err = s.roleRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取角色失败: %w", err)
	}

	return s.toRoleVO(role), nil
}

// DeleteRole 删除角色
func (s *RoleService) DeleteRole(id uint) error {
	// 检查角色是否存在
	_, err := s.roleRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("角色不存在: %w", err)
	}

	return s.roleRepo.Delete(id)
}

// GetRole 获取角色详情
func (s *RoleService) GetRole(id uint) (*RoleVO, error) {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("角色不存在")
		}
		return nil, fmt.Errorf("获取角色失败: %w", err)
	}
	return s.toRoleVO(role), nil
}

// ListRoles 角色列表
func (s *RoleService) ListRoles(page, pageSize int, name string) ([]*RoleVO, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	roles, total, err := s.roleRepo.List(page, pageSize, name)
	if err != nil {
		return nil, 0, fmt.Errorf("获取角色列表失败: %w", err)
	}

	vos := make([]*RoleVO, 0, len(roles))
	for i := range roles {
		vos = append(vos, s.toRoleVO(&roles[i]))
	}

	return vos, total, nil
}

// ListAllRoles 获取所有角色（不分页）
func (s *RoleService) ListAllRoles() ([]*RoleVO, error) {
	roles, err := s.roleRepo.ListAll()
	if err != nil {
		return nil, fmt.Errorf("获取角色列表失败: %w", err)
	}

	vos := make([]*RoleVO, 0, len(roles))
	for i := range roles {
		vos = append(vos, s.toRoleVO(&roles[i]))
	}

	return vos, nil
}

// AssignPermissions 分配权限
func (s *RoleService) AssignPermissions(roleID uint, permissionIDs []uint) error {
	// 检查角色是否存在
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return fmt.Errorf("角色不存在: %w", err)
	}

	return s.roleRepo.AssignPermissions(roleID, permissionIDs)
}

// ListPermissions 获取所有权限
func (s *RoleService) ListPermissions() ([]*PermissionVO, error) {
	permissions, err := s.permissionRepo.List()
	if err != nil {
		return nil, fmt.Errorf("获取权限列表失败: %w", err)
	}

	vos := make([]*PermissionVO, 0, len(permissions))
	for i := range permissions {
		vos = append(vos, s.toPermissionVO(&permissions[i]))
	}

	return vos, nil
}

// HasPermission 检查角色是否有指定权限
func (s *RoleService) HasPermission(roleID uint, resource, action string) bool {
	permissions, err := s.permissionRepo.GetPermissionsByRoleID(roleID)
	if err != nil {
		return false
	}

	for _, p := range permissions {
		if p.Resource == resource && p.Action == action {
			return true
		}
		// 支持 * 通配符
		if p.Resource == "*" || p.Action == "*" {
			return true
		}
		if p.Resource == resource && p.Action == "*" {
			return true
		}
		if p.Resource == "*" && p.Action == action {
			return true
		}
	}

	return false
}

// toRoleVO 转换为 RoleVO
func (s *RoleService) toRoleVO(role *model.Role) *RoleVO {
	permissions := make([]PermissionVO, 0, len(role.Permissions))
	for _, p := range role.Permissions {
		permissions = append(permissions, PermissionVO{
			ID:       p.ID,
			Name:     p.Name,
			Resource: p.Resource,
			Action:   p.Action,
			MenuKey:  p.MenuKey,
		})
	}

	return &RoleVO{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: permissions,
		CreatedAt:   role.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   role.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// toPermissionVO 转换为 PermissionVO
func (s *RoleService) toPermissionVO(p *model.Permission) *PermissionVO {
	return &PermissionVO{
		ID:       p.ID,
		Name:     p.Name,
		Resource: p.Resource,
		Action:   p.Action,
		MenuKey:  p.MenuKey,
	}
}
