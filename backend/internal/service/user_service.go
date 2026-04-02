package service

import (
	"errors"
	"fmt"

	"certmanager-backend/internal/model"
	"certmanager-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务层
type UserService struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
}

// NewUserService 创建 UserService 实例
func NewUserService(userRepo *repository.UserRepository, roleRepo *repository.RoleRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	RoleID   uint   `json:"role_id"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Status   string `json:"status"`
	RoleID   uint   `json:"role_id"`
}

// CreateUser 创建用户
func (s *UserService) CreateUser(req *CreateUserRequest) (*UserVO, error) {
	// 检查用户名是否存在
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否存在
	if req.Email != "" {
		exists, err = s.userRepo.ExistsByEmail(req.Email)
		if err != nil {
			return nil, fmt.Errorf("检查邮箱失败: %w", err)
		}
		if exists {
			return nil, errors.New("邮箱已存在")
		}
	}

	// 检查角色是否存在
	if req.RoleID > 0 {
		_, err := s.roleRepo.GetByID(req.RoleID)
		if err != nil {
			return nil, errors.New("角色不存在")
		}
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码哈希失败: %w", err)
	}

	user := &model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Nickname: req.Nickname,
		RoleID:   req.RoleID,
		Status:   "active",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 重新获取用户（包含角色信息）
	user, err = s.userRepo.GetByID(user.ID)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	return s.toUserVO(user), nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(id uint, req *UpdateUserRequest) (*UserVO, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 检查邮箱是否已被其他用户使用
	if req.Email != "" && req.Email != user.Email {
		exists, err := s.userRepo.ExistsByEmail(req.Email)
		if err != nil {
			return nil, fmt.Errorf("检查邮箱失败: %w", err)
		}
		if exists {
			return nil, errors.New("邮箱已被使用")
		}
		user.Email = req.Email
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Status != "" {
		user.Status = req.Status
	}
	if req.RoleID > 0 {
		// 检查角色是否存在
		_, err := s.roleRepo.GetByID(req.RoleID)
		if err != nil {
			return nil, errors.New("角色不存在")
		}
		user.RoleID = req.RoleID
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	// 重新获取用户（包含角色信息）
	user, err = s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	return s.toUserVO(user), nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	return s.userRepo.Delete(id)
}

// GetUser 获取用户详情
func (s *UserService) GetUser(id uint) (*UserVO, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	return s.toUserVO(user), nil
}

// ListUsers 用户列表
func (s *UserService) ListUsers(page, pageSize int, username, status string) ([]*UserVO, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := s.userRepo.List(page, pageSize, username, status)
	if err != nil {
		return nil, 0, fmt.Errorf("获取用户列表失败: %w", err)
	}

	vos := make([]*UserVO, 0, len(users))
	for i := range users {
		vos = append(vos, s.toUserVO(&users[i]))
	}

	return vos, total, nil
}

// AssignRole 分配角色
func (s *UserService) AssignRole(userID, roleID uint) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 检查角色是否存在
	_, err = s.roleRepo.GetByID(roleID)
	if err != nil {
		return errors.New("角色不存在")
	}

	return s.userRepo.AssignRole(userID, roleID)
}

// ResetPassword 重置密码
func (s *UserService) ResetPassword(userID uint, newPassword string) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %w", err)
	}

	return s.userRepo.UpdatePassword(userID, string(hashedPassword))
}

// toUserVO 转换为 UserVO
func (s *UserService) toUserVO(user *model.User) *UserVO {
	roleName := ""
	if user.Role.ID != 0 {
		roleName = user.Role.Name
	}

	lastLoginAt := ""
	if user.LastLoginAt != nil {
		lastLoginAt = user.LastLoginAt.Format("2006-01-02 15:04:05")
	}

	return &UserVO{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Nickname:    user.Nickname,
		RoleID:      user.RoleID,
		RoleName:    roleName,
		Status:      user.Status,
		LastLoginAt: lastLoginAt,
		CreatedAt:   user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
