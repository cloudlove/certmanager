package handler

import (
	"net/http"
	"strconv"

	"certmanager-backend/internal/service"
	"certmanager-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// CSRHandler CSR HTTP Handler
type CSRHandler struct {
	svc *service.CSRService
}

// NewCSRHandler 创建 CSRHandler 实例
func NewCSRHandler(svc *service.CSRService) *CSRHandler {
	return &CSRHandler{svc: svc}
}

// RegisterRoutes 注册 CSR 路由
func (h *CSRHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/generate", h.Generate)
	rg.GET("", h.List)
	rg.GET("/:id", h.Get)
	rg.DELETE("/:id", h.Delete)
	rg.POST("/parse", h.Parse)
	rg.GET("/:id/download-csr", h.DownloadCSR)
	rg.GET("/:id/download-key", h.DownloadPrivateKey)
}

// generateCSRReq 生成 CSR 请求
type generateCSRReq struct {
	CommonName   string   `json:"common_name" binding:"required"`
	SANs         []string `json:"sans"`
	KeyAlgorithm string   `json:"key_algorithm" binding:"required"`
	KeySize      int      `json:"key_size" binding:"required"`
	// 新增阿里云 CreateCsr 参数
	CountryCode string `json:"country_code" binding:"required"`
	Province    string `json:"province" binding:"required"`
	Locality    string `json:"locality" binding:"required"`
	CorpName    string `json:"corp_name"`
	Department  string `json:"department"`
}

// parseCSRReq 解析 CSR 请求
type parseCSRReq struct {
	CSRPEM string `json:"csr_pem" binding:"required"`
}

// Generate 生成 CSR
// POST /api/v1/csrs/generate
func (h *CSRHandler) Generate(c *gin.Context) {
	var req generateCSRReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	vo, err := h.svc.Generate(req.CommonName, req.SANs, req.KeyAlgorithm, req.KeySize, req.CountryCode, req.Province, req.Locality, req.CorpName, req.Department)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, vo)
}

// List CSR 列表
// GET /api/v1/csrs?page=1&pageSize=10&search=example
func (h *CSRHandler) List(c *gin.Context) {
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

// Get 获取 CSR 详情
// GET /api/v1/csrs/:id
func (h *CSRHandler) Get(c *gin.Context) {
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

// Delete 删除 CSR
// DELETE /api/v1/csrs/:id
func (h *CSRHandler) Delete(c *gin.Context) {
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

// Parse 解析 CSR
// POST /api/v1/csrs/parse
func (h *CSRHandler) Parse(c *gin.Context) {
	var req parseCSRReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	info, err := h.svc.Parse(req.CSRPEM)
	if err != nil {
		response.Error(c, 400, "failed to parse CSR: "+err.Error())
		return
	}

	response.Success(c, info)
}

// DownloadCSR 下载 CSR
// GET /api/v1/csrs/:id/download-csr
func (h *CSRHandler) DownloadCSR(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	csrPEM, err := h.svc.DownloadCSR(id)
	if err != nil {
		response.Error(c, 404, err.Error())
		return
	}

	c.Header("Content-Type", "application/x-pem-file")
	c.Header("Content-Disposition", "attachment; filename=csr.pem")
	c.String(http.StatusOK, csrPEM)
}

// DownloadPrivateKey 下载私钥
// GET /api/v1/csrs/:id/download-key
func (h *CSRHandler) DownloadPrivateKey(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	privateKeyPEM, err := h.svc.DownloadPrivateKey(id)
	if err != nil {
		response.Error(c, 404, err.Error())
		return
	}

	c.Header("Content-Type", "application/x-pem-file")
	c.Header("Content-Disposition", "attachment; filename=private_key.pem")
	c.String(http.StatusOK, privateKeyPEM)
}
