package handler

import (
	"strconv"

	"certmanager-backend/internal/service"
	"certmanager-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// CredentialHandler 云凭证 HTTP Handler
type CredentialHandler struct {
	svc *service.CredentialService
}

// NewCredentialHandler 创建 CredentialHandler 实例
func NewCredentialHandler(svc *service.CredentialService) *CredentialHandler {
	return &CredentialHandler{svc: svc}
}

// RegisterRoutes 注册凭证路由
func (h *CredentialHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", h.Create)
	rg.GET("", h.List)
	rg.GET("/:id", h.Get)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
	rg.POST("/:id/test", h.TestConnection)
}

// createCredentialReq 创建凭证请求
type createCredentialReq struct {
	Name         string `json:"name" binding:"required"`
	ProviderType string `json:"provider_type" binding:"required"`
	AccessKey    string `json:"access_key" binding:"required"`
	SecretKey    string `json:"secret_key" binding:"required"`
	ExtraConfig  string `json:"extra_config"`
}

// updateCredentialReq 更新凭证请求
type updateCredentialReq struct {
	Name         string `json:"name"`
	ProviderType string `json:"provider_type"`
	AccessKey    string `json:"access_key"`
	SecretKey    string `json:"secret_key"`
	ExtraConfig  string `json:"extra_config"`
}

// Create 创建凭证
// POST /api/v1/credentials
func (h *CredentialHandler) Create(c *gin.Context) {
	var req createCredentialReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	vo, err := h.svc.Create(req.Name, req.ProviderType, req.AccessKey, req.SecretKey, req.ExtraConfig)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, vo)
}

// Update 更新凭证
// PUT /api/v1/credentials/:id
func (h *CredentialHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	var req updateCredentialReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	vo, err := h.svc.Update(id, req.Name, req.ProviderType, req.AccessKey, req.SecretKey, req.ExtraConfig)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, vo)
}

// Delete 删除凭证
// DELETE /api/v1/credentials/:id
func (h *CredentialHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, nil)
}

// Get 获取凭证详情
// GET /api/v1/credentials/:id
func (h *CredentialHandler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	vo, err := h.svc.Get(id)
	if err != nil {
		response.Error(c, 404, err.Error())
		return
	}

	response.Success(c, vo)
}

// List 凭证列表
// GET /api/v1/credentials?page=1&pageSize=10&providerType=aliyun
func (h *CredentialHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	providerType := c.Query("providerType")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	vos, total, err := h.svc.List(page, pageSize, providerType)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.SuccessWithPage(c, vos, total, page, pageSize)
}

// TestConnection 测试凭证连通性
// POST /api/v1/credentials/:id/test
func (h *CredentialHandler) TestConnection(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	status, msg := h.svc.TestConnection(id)
	response.Success(c, gin.H{
		"status":  status,
		"message": msg,
	})
}

// parseID 从路径参数解析 uint ID
func parseID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
