package handler

import (
	"strconv"

	"certmanager-backend/internal/model"
	"certmanager-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuditHandler 审计日志 Handler
type AuditHandler struct {
	db *gorm.DB
}

// NewAuditHandler 创建 AuditHandler 实例
func NewAuditHandler(db *gorm.DB) *AuditHandler {
	return &AuditHandler{db: db}
}

// RegisterRoutes 注册审计日志路由
func (h *AuditHandler) RegisterRoutes(rg *gin.RouterGroup) {
	audit := rg.Group("/audit-logs")
	{
		audit.GET("", h.ListAuditLogs)
	}
}

// ListAuditLogs 获取审计日志列表
// GET /api/v1/audit-logs?page=1&pageSize=10&resourceType=&action=
func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	resourceType := c.Query("resourceType")
	action := c.Query("action")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询
	query := h.db.Model(&model.AuditLog{})
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, 500, "failed to count audit logs")
		return
	}

	// 获取数据
	var logs []model.AuditLog
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		response.Error(c, 500, "failed to query audit logs")
		return
	}

	response.SuccessWithPage(c, logs, total, page, pageSize)
}
