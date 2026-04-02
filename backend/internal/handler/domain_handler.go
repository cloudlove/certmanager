package handler

import (
	"strconv"

	"certmanager-backend/internal/service"
	"certmanager-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// DomainHandler 域名 HTTP Handler
type DomainHandler struct {
	svc *service.DomainService
}

// NewDomainHandler 创建 DomainHandler 实例
func NewDomainHandler(svc *service.DomainService) *DomainHandler {
	return &DomainHandler{svc: svc}
}

// RegisterRoutes 注册域名路由
func (h *DomainHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", h.Create)
	rg.GET("", h.List)
	rg.GET("/:id", h.Get)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
	rg.POST("/:id/verify", h.Verify)
	rg.POST("/batch-verify", h.BatchVerify)
}

// createDomainReq 创建域名请求
type createDomainReq struct {
	Name string `json:"name" binding:"required"`
}

// updateDomainReq 更新域名请求
type updateDomainReq struct {
	CertificateID *uint `json:"certificate_id"`
}

// batchVerifyReq 批量校验请求
type batchVerifyReq struct {
	IDs []uint `json:"ids" binding:"required"`
}

// Create 创建域名
// POST /api/v1/domains
func (h *DomainHandler) Create(c *gin.Context) {
	var req createDomainReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	vo, err := h.svc.Create(req.Name)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, vo)
}

// Update 更新域名（关联/取消关联证书）
// PUT /api/v1/domains/:id
func (h *DomainHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	var req updateDomainReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	vo, err := h.svc.Update(id, req.CertificateID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, vo)
}

// Delete 删除域名
// DELETE /api/v1/domains/:id
func (h *DomainHandler) Delete(c *gin.Context) {
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

// Get 获取域名详情
// GET /api/v1/domains/:id
func (h *DomainHandler) Get(c *gin.Context) {
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

// List 域名列表
// GET /api/v1/domains?page=1&pageSize=10&search=example
func (h *DomainHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.Query("search")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	vos, total, err := h.svc.List(page, pageSize, search)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.SuccessWithPage(c, vos, total, page, pageSize)
}

// Verify 校验单个域名证书
// POST /api/v1/domains/:id/verify
func (h *DomainHandler) Verify(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	result, err := h.svc.VerifyDomainCert(id)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, result)
}

// BatchVerify 批量校验域名证书
// POST /api/v1/domains/batch-verify
func (h *DomainHandler) BatchVerify(c *gin.Context) {
	var req batchVerifyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	if len(req.IDs) == 0 {
		response.Error(c, 400, "ids is required")
		return
	}

	results, err := h.svc.BatchVerify(req.IDs)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, gin.H{
		"results": results,
	})
}
