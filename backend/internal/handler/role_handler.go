package handler

import (
	"strconv"

	"certmanager-backend/internal/service"
	"certmanager-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// RoleHandler 角色管理 HTTP Handler
type RoleHandler struct {
	roleSvc *service.RoleService
}

// NewRoleHandler 创建 RoleHandler 实例
func NewRoleHandler(roleSvc *service.RoleService) *RoleHandler {
	return &RoleHandler{roleSvc: roleSvc}
}

// RegisterRoutes 注册角色管理路由
func (h *RoleHandler) RegisterRoutes(rg *gin.RouterGroup) {
	roles := rg.Group("/roles")
	{
		roles.GET("", h.List)
		roles.POST("", h.Create)
		roles.GET("/all", h.ListAll)
		roles.GET("/permissions", h.ListPermissions)
		roles.GET("/:id", h.Get)
		roles.PUT("/:id", h.Update)
		roles.DELETE("/:id", h.Delete)
		roles.PUT("/:id/permissions", h.AssignPermissions)
	}
}

// createRoleReq 创建角色请求
type createRoleReq struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	PermissionIDs []uint `json:"permission_ids"`
}

// updateRoleReq 更新角色请求
type updateRoleReq struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	PermissionIDs []uint `json:"permission_ids"`
}

// assignPermissionsReq 分配权限请求
type assignPermissionsReq struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

// Create 创建角色
// POST /api/v1/roles
func (h *RoleHandler) Create(c *gin.Context) {
	var req createRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "无效的请求: "+err.Error())
		return
	}

	role, err := h.roleSvc.CreateRole(&service.CreateRoleRequest{
		Name:          req.Name,
		Description:   req.Description,
		PermissionIDs: req.PermissionIDs,
	})
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, role)
}

// Update 更新角色
// PUT /api/v1/roles/:id
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Error(c, 400, "无效的角色ID")
		return
	}

	var req updateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "无效的请求: "+err.Error())
		return
	}

	role, err := h.roleSvc.UpdateRole(id, &service.UpdateRoleRequest{
		Name:          req.Name,
		Description:   req.Description,
		PermissionIDs: req.PermissionIDs,
	})
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, role)
}

// Delete 删除角色
// DELETE /api/v1/roles/:id
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Error(c, 400, "无效的角色ID")
		return
	}

	if err := h.roleSvc.DeleteRole(id); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, nil)
}

// Get 获取角色详情
// GET /api/v1/roles/:id
func (h *RoleHandler) Get(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Error(c, 400, "无效的角色ID")
		return
	}

	role, err := h.roleSvc.GetRole(id)
	if err != nil {
		response.Error(c, 404, err.Error())
		return
	}

	response.Success(c, role)
}

// List 角色列表
// GET /api/v1/roles?page=1&pageSize=10&name=xxx
func (h *RoleHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	name := c.Query("name")

	roles, total, err := h.roleSvc.ListRoles(page, pageSize, name)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.SuccessWithPage(c, roles, total, page, pageSize)
}

// ListAll 获取所有角色（不分页）
// GET /api/v1/roles/all
func (h *RoleHandler) ListAll(c *gin.Context) {
	roles, err := h.roleSvc.ListAllRoles()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, roles)
}

// AssignPermissions 分配权限
// PUT /api/v1/roles/:id/permissions
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Error(c, 400, "无效的角色ID")
		return
	}

	var req assignPermissionsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "无效的请求: "+err.Error())
		return
	}

	if err := h.roleSvc.AssignPermissions(id, req.PermissionIDs); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "权限分配成功"})
}

// ListPermissions 获取所有权限列表
// GET /api/v1/roles/permissions
func (h *RoleHandler) ListPermissions(c *gin.Context) {
	permissions, err := h.roleSvc.ListPermissions()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, permissions)
}
