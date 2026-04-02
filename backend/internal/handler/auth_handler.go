package handler

import (
	"certmanager-backend/internal/middleware"
	"certmanager-backend/internal/service"
	"certmanager-backend/pkg/password"
	"certmanager-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证 HTTP Handler
type AuthHandler struct {
	authSvc *service.AuthService
}

// NewAuthHandler 创建 AuthHandler 实例
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// RegisterRoutes 注册认证路由（公开路由）
func (h *AuthHandler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
	}
}

// RegisterProtectedRoutes 注册受保护的认证路由
func (h *AuthHandler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.GET("/me", h.GetCurrentUser)
		auth.PUT("/password", h.ChangePassword)
		auth.POST("/logout", h.Logout)
	}
}

// loginReq 登录请求
type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 用户登录
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "无效的请求: "+err.Error())
		return
	}

	resp, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		response.Error(c, 401, err.Error())
		return
	}

	response.Success(c, resp)
}

// refreshReq 刷新令牌请求
type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken 刷新令牌
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "无效的请求: "+err.Error())
		return
	}

	resp, err := h.authSvc.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(c, 401, err.Error())
		return
	}

	response.Success(c, resp)
}

// GetCurrentUser 获取当前用户信息
// GET /api/v1/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, _, _, ok := middleware.GetCurrentUser(c)
	if !ok {
		response.Error(c, 401, "未授权")
		return
	}

	user, err := h.authSvc.GetCurrentUser(userID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, user)
}

// changePasswordReq 修改密码请求
type changePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword 修改密码
// PUT /api/v1/auth/password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, _, _, ok := middleware.GetCurrentUser(c)
	if !ok {
		response.Error(c, 401, "未授权")
		return
	}

	var req changePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "无效的请求: "+err.Error())
		return
	}

	// 验证新密码强度
	if err := password.ValidatePasswordStrength(req.NewPassword); err != nil {
		response.Error(c, 400, "密码强度不足: "+err.Error())
		return
	}

	if err := h.authSvc.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "密码修改成功"})
}

// Logout 用户登出
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// 简单实现：客户端清除 token 即可
	// 如需实现 token 黑名单，可以使用 Redis 存储
	response.Success(c, gin.H{"message": "登出成功"})
}
