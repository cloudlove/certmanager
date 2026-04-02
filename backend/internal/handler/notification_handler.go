package handler

import (
	"strconv"

	"certmanager-backend/internal/service"
	"certmanager-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// NotificationHandler 通知 HTTP Handler
type NotificationHandler struct {
	svc *service.NotificationService
}

// NewNotificationHandler 创建 NotificationHandler 实例
func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// RegisterRoutes 注册通知路由
func (h *NotificationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// 通知规则路由
	rg.POST("/rules", h.CreateRule)
	rg.GET("/rules", h.ListRules)
	rg.GET("/rules/:id", h.GetRule)
	rg.PUT("/rules/:id", h.UpdateRule)
	rg.DELETE("/rules/:id", h.DeleteRule)
	rg.PUT("/rules/:id/toggle", h.ToggleRule)
	rg.POST("/rules/:id/test", h.TestRule)

	// 通知日志路由
	rg.GET("/logs", h.ListLogs)
}

// createRuleReq 创建通知规则请求
type createRuleReq struct {
	Name          string   `json:"name" binding:"required"`
	EventType     string   `json:"event_type" binding:"required"`
	ThresholdDays int      `json:"threshold_days"`
	Channels      []string `json:"channels" binding:"required"`
	Recipients    []string `json:"recipients" binding:"required"`
	Enabled       bool     `json:"enabled"`
}

// updateRuleReq 更新通知规则请求
type updateRuleReq struct {
	Name          string   `json:"name"`
	EventType     string   `json:"event_type"`
	ThresholdDays int      `json:"threshold_days"`
	Channels      []string `json:"channels"`
	Recipients    []string `json:"recipients"`
	Enabled       bool     `json:"enabled"`
}

// toggleRuleReq 切换规则状态请求
type toggleRuleReq struct {
	Enabled bool `json:"enabled" binding:"required"`
}

// CreateRule 创建通知规则
// POST /api/v1/notification-rules
func (h *NotificationHandler) CreateRule(c *gin.Context) {
	var req createRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	createReq := &service.CreateRuleRequest{
		Name:          req.Name,
		EventType:     req.EventType,
		ThresholdDays: req.ThresholdDays,
		Channels:      req.Channels,
		Recipients:    req.Recipients,
		Enabled:       req.Enabled,
	}

	vo, err := h.svc.CreateRule(createReq)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, vo)
}

// UpdateRule 更新通知规则
// PUT /api/v1/notification-rules/:id
func (h *NotificationHandler) UpdateRule(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	var req updateRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	updateReq := &service.UpdateRuleRequest{
		Name:          req.Name,
		EventType:     req.EventType,
		ThresholdDays: req.ThresholdDays,
		Channels:      req.Channels,
		Recipients:    req.Recipients,
		Enabled:       req.Enabled,
	}

	vo, err := h.svc.UpdateRule(id, updateReq)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, vo)
}

// DeleteRule 删除通知规则
// DELETE /api/v1/notification-rules/:id
func (h *NotificationHandler) DeleteRule(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	if err := h.svc.DeleteRule(id); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetRule 获取通知规则详情
// GET /api/v1/notification-rules/:id
func (h *NotificationHandler) GetRule(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	vo, err := h.svc.GetRule(id)
	if err != nil {
		response.Error(c, 404, err.Error())
		return
	}

	response.Success(c, vo)
}

// ListRules 获取通知规则列表
// GET /api/v1/notification-rules?page=1&pageSize=10
func (h *NotificationHandler) ListRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	vos, total, err := h.svc.ListRules(page, pageSize)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.SuccessWithPage(c, vos, total, page, pageSize)
}

// ToggleRule 切换通知规则启用状态
// PUT /api/v1/notification-rules/:id/toggle
func (h *NotificationHandler) ToggleRule(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	var req toggleRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	if err := h.svc.ToggleRule(id, req.Enabled); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, gin.H{"enabled": req.Enabled})
}

// TestRule 测试通知规则
// POST /api/v1/notification-rules/:id/test
func (h *NotificationHandler) TestRule(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	if err := h.svc.TestRule(id); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "test notification sent"})
}

// ListLogs 获取通知日志列表
// GET /api/v1/notification-logs?page=1&pageSize=10&eventType=cert_expiry
func (h *NotificationHandler) ListLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	eventType := c.Query("eventType")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	vos, total, err := h.svc.ListLogs(page, pageSize, eventType)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.SuccessWithPage(c, vos, total, page, pageSize)
}
