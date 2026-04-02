package handler

import (
	"net/http"
	"strconv"

	"certmanager-backend/internal/service"
	"certmanager-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// CertificateHandler 证书 HTTP Handler
type CertificateHandler struct {
	svc *service.CertificateService
}

// NewCertificateHandler 创建 CertificateHandler 实例
func NewCertificateHandler(svc *service.CertificateService) *CertificateHandler {
	return &CertificateHandler{svc: svc}
}

// RegisterRoutes 注册证书路由
func (h *CertificateHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/apply", h.Apply)                        // 申请证书
	rg.POST("/import", h.Import)                      // 导入证书
	rg.GET("", h.List)                                // 证书列表
	rg.GET("/:id", h.Get)                             // 获取证书详情
	rg.DELETE("/:id", h.Delete)                       // 删除证书
	rg.POST("/:id/sync", h.Sync)                      // 同步证书状态
	rg.GET("/:id/download-cert", h.DownloadCert)      // 下载证书
	rg.GET("/:id/download-key", h.DownloadPrivateKey) // 下载私钥
}

// applyCertReq 申请证书请求
type applyCertReq struct {
	CAProvider   string `json:"ca_provider" binding:"required"`
	Domain       string `json:"domain" binding:"required"`
	CSRID        uint   `json:"csr_id" binding:"required"`
	CredentialID uint   `json:"credential_id" binding:"required"`
	ValidateType string `json:"validate_type" binding:"required"` // DNS 或 FILE
	ProductType  string `json:"product_type" binding:"required"`  // DV / OV / EV
	DomainType   string `json:"domain_type" binding:"required"`   // single / wildcard / multi
}

// importCertReq 导入证书请求
type importCertReq struct {
	CertPEM       string `json:"cert_pem" binding:"required"`
	ChainPEM      string `json:"chain_pem"`
	PrivateKeyPEM string `json:"private_key_pem"`
}

// Apply 申请证书
// POST /api/v1/certificates/apply
func (h *CertificateHandler) Apply(c *gin.Context) {
	var req applyCertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	vo, err := h.svc.ApplyCertificate(req.CAProvider, req.Domain, req.CSRID, req.CredentialID, req.ValidateType, req.ProductType, req.DomainType)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, vo)
}

// Import 导入证书
// POST /api/v1/certificates/import
func (h *CertificateHandler) Import(c *gin.Context) {
	var req importCertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	vo, err := h.svc.ImportCertificate(req.CertPEM, req.ChainPEM, req.PrivateKeyPEM)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, vo)
}

// List 证书列表
// GET /api/v1/certificates?page=1&pageSize=10&status=&search=&sortBy=
func (h *CertificateHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")
	search := c.Query("search")
	sortBy := c.Query("sortBy")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	vos, total, err := h.svc.List(page, pageSize, status, search, sortBy)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.SuccessWithPage(c, vos, total, page, pageSize)
}

// Get 获取证书详情
// GET /api/v1/certificates/:id
func (h *CertificateHandler) Get(c *gin.Context) {
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

// Delete 删除证书
// DELETE /api/v1/certificates/:id
func (h *CertificateHandler) Delete(c *gin.Context) {
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

// Sync 同步证书状态
// POST /api/v1/certificates/:id/sync
func (h *CertificateHandler) Sync(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	vo, err := h.svc.SyncCertStatus(id)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, vo)
}

// DownloadCert 下载证书
// GET /api/v1/certificates/:id/download-cert
func (h *CertificateHandler) DownloadCert(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	certPEM, chainPEM, err := h.svc.DownloadCertificate(id)
	if err != nil {
		response.Error(c, 404, err.Error())
		return
	}

	// 合并证书和证书链
	fullCert := certPEM
	if chainPEM != "" {
		fullCert += "\n" + chainPEM
	}

	c.Header("Content-Type", "application/x-pem-file")
	c.Header("Content-Disposition", "attachment; filename=certificate.pem")
	c.String(http.StatusOK, fullCert)
}

// DownloadPrivateKey 下载私钥
// GET /api/v1/certificates/:id/download-key
func (h *CertificateHandler) DownloadPrivateKey(c *gin.Context) {
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
