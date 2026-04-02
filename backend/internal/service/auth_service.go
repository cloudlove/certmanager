package service

import (
	"errors"
	"fmt"
	"time"

	"certmanager-backend/internal/config"
	"certmanager-backend/internal/model"
	"certmanager-backend/internal/repository"
	"certmanager-backend/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务层
type AuthService struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
}

// NewAuthService 创建 AuthService 实例
func NewAuthService(userRepo *repository.UserRepository, roleRepo *repository.RoleRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         *UserVO   `json:"user"`
}

// UserVO 用户视图对象
type UserVO struct {
	ID          uint     `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Nickname    string   `json:"nickname"`
	RoleID      uint     `json:"role_id"`
	RoleName    string   `json:"role_name"`
	Role        string   `json:"role"`        // 角色标识（admin/operator/viewer）
	Permissions []string `json:"permissions"` // 权限列表 ["resource:action", ...]
	Status      string   `json:"status"`
	LastLoginAt string   `json:"last_login_at"`
	CreatedAt   string   `json:"created_at"`
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (*LoginResponse, error) {
	// 获取用户
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("用户已被禁用")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 获取配置
	cfg := config.GetConfig()

	// 生成 token
	accessToken, refreshToken, err := jwt.GenerateTokenPair(
		user.ID,
		user.Username,
		user.RoleID,
		cfg.JWT.Secret,
		cfg.JWT.ExpireHours,
		cfg.JWT.RefreshHours,
	)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	// 更新最后登录时间
	_ = s.userRepo.UpdateLastLogin(user.ID)

	// 计算过期时间
	expiresAt := time.Now().Add(time.Duration(cfg.JWT.ExpireHours) * time.Hour)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         s.toUserVO(user),
	}, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(refreshToken string) (*LoginResponse, error) {
	// 获取配置
	cfg := config.GetConfig()

	// 解析 refresh token
	claims, err := jwt.ParseToken(refreshToken, cfg.JWT.Secret)
	if err != nil {
		if err == jwt.ErrExpiredToken {
			return nil, errors.New("刷新令牌已过期")
		}
		return nil, errors.New("无效的刷新令牌")
	}

	// 检查 token 类型
	if claims.TokenType != jwt.TokenTypeRefresh {
		return nil, errors.New("无效的令牌类型")
	}

	// 获取用户
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("用户已被禁用")
	}

	// 生成新的 token
	accessToken, newRefreshToken, err := jwt.GenerateTokenPair(
		user.ID,
		user.Username,
		user.RoleID,
		cfg.JWT.Secret,
		cfg.JWT.ExpireHours,
		cfg.JWT.RefreshHours,
	)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	// 计算过期时间
	expiresAt := time.Now().Add(time.Duration(cfg.JWT.ExpireHours) * time.Hour)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		User:         s.toUserVO(user),
	}, nil
}

// GetCurrentUser 获取当前用户信息
func (s *AuthService) GetCurrentUser(userID uint) (*UserVO, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	return s.toUserVO(user), nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	// 获取用户
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("旧密码错误")
	}

	// 哈希新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %w", err)
	}

	// 更新密码
	if err := s.userRepo.UpdatePassword(userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}

// HashPassword 哈希密码
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// toUserVO 转换为 UserVO
func (s *AuthService) toUserVO(user *model.User) *UserVO {
	roleName := ""
	roleCode := ""
	if user.Role.ID != 0 {
		roleName = user.Role.Name
		roleCode = user.Role.Name // 使用角色名称作为角色标识
	}

	// 转换权限列表为 ["resource:action"] 格式
	var permissions []string
	if len(user.Role.Permissions) > 0 {
		permissions = make([]string, 0, len(user.Role.Permissions))
		for _, perm := range user.Role.Permissions {
			permissions = append(permissions, perm.Resource+":"+perm.Action)
		}
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
		Role:        roleCode,
		Permissions: permissions,
		Status:      user.Status,
		LastLoginAt: lastLoginAt,
		CreatedAt:   user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
