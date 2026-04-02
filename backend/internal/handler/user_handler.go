package handler

import (
	"strconv"

	"certmanager-backend/internal/middleware"
	"certmanager-backend/internal/service"
	"certmanager-backend/pkg/password"
	"certmanager-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户管理 HTTP Handler
type UserHandler struct {
	userSvc *service.UserService
}

// NewUserHandler 创建 UserHandler 实例
func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// RegisterRoutes 注册用户管理路由
func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.GET("", h.List)
		users.POST("", h.Create)
		users.GET("/:id", h.Get)
		users.PUT("/:id", h.Update)
		users.DELETE("/:id", h.Delete)
		users.PUT("/:id/role", h.AssignRole)
		users.PUT("/:id/password", h.ResetPassword)
	}
}

// createUserReq 创建用户请求
type createUserReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	RoleID   uint   `json:"role_id"`
}

// updateUserReq 更新用户请求
type updateUserReq struct {
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Status   string `json:"status"`
	RoleID   uint   `json:"role_id"`
}

// assignRoleReq 分配角色请求
type assignRoleReq struct {
	RoleID uint `json:"role_id" binding:"required"`
}

// resetPasswordReq 重置密码请求
type resetPasswordReq struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// Create 创建用户
// POST /api/v1/users
func (h *UserHandler) Create(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "无效的请求: "+err.Error())
		return
	}

	// 验证密码强度
	if err := password.ValidatePasswordStrength(req.Password); err != nil {
		response.Error(c, 400, "密码强度不足: "+err.Error())
		return
	}

	user, err := h.userSvc.CreateUser(&service.CreateUserRequest{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Nickname: req.Nickname,
		RoleID:   req.RoleID,
	})
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, user)
}

// Update 更新用户
// PUT /api/v1/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Error(c, 400, "无效的用户ID")
		return
	}

	var req updateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "无效的请求: "+err.Error())
		return
	}

	user, err := h.userSvc.UpdateUser(id, &service.UpdateUserRequest{
		Email:    req.Email,
		Nickname: req.Nickname,
		Status:   req.Status,
		RoleID:   req.RoleID,
	})
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, user)
}

// Delete 删除用户
// DELETE /api/v1/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Error(c, 400, "无效的用户ID")
		return
	}

	// 检查是否删除自己
	userID, _, _, ok := middleware.GetCurrentUser(c)
	if ok && userID == id {
		response.Error(c, 400, "不能删除自己")
		return
	}

	if err := h.userSvc.DeleteUser(id); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, nil)
}

// Get 获取用户详情
// GET /api/v1/users/:id
func (h *UserHandler) Get(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Error(c, 400, "无效的用户ID")
		return
	}

	user, err := h.userSvc.GetUser(id)
	if err != nil {
		response.Error(c, 404, err.Error())
		return
	}

	response.Success(c, user)
}

// List 用户列表
// GET /api/v1/users?page=1&pageSize=10&username=xxx&status=active
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	username := c.Query("username")
	status := c.Query("status")

	users, total, err := h.userSvc.ListUsers(page, pageSize, username, status)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.SuccessWithPage(c, users, total, page, pageSize)
}

// AssignRole 分配角色
// PUT /api/v1/users/:id/role
func (h *UserHandler) AssignRole(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Error(c, 400, "无效的用户ID")
		return
	}

	var req assignRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "无效的请求: "+err.Error())
		return
	}

	if err := h.userSvc.AssignRole(id, req.RoleID); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "角色分配成功"})
}

// ResetPassword 重置密码
// PUT /api/v1/users/:id/password
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Error(c, 400, "无效的用户ID")
		return
	}

	var req resetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "无效的请求: "+err.Error())
		return
	}

	// 验证密码强度
	if err := password.ValidatePasswordStrength(req.NewPassword); err != nil {
		response.Error(c, 400, "密码强度不足: "+err.Error())
		return
	}

	if err := h.userSvc.ResetPassword(id, req.NewPassword); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "密码重置成功"})
}

// parseIDParam 从路径参数解析 uint ID
func parseIDParam(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
