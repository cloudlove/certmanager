package handler

import (
	"net/http"
	"strconv"

	"certmanager-backend/internal/service"
	"certmanager-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// DashboardHandler 大盘统计 HTTP Handler
type DashboardHandler struct {
	svc *service.DashboardService
}

// NewDashboardHandler 创建 DashboardHandler 实例
func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// RegisterRoutes 注册大盘路由
func (h *DashboardHandler) RegisterRoutes(rg *gin.RouterGroup) {
	dashboard := rg.Group("/dashboard")
	{
		dashboard.GET("/overview", h.GetOverview)
		dashboard.GET("/cert-overview", h.GetCertOverview)
		dashboard.GET("/deploy-overview", h.GetDeployOverview)
		dashboard.GET("/cloud-distribution", h.GetCloudDistribution)
		dashboard.GET("/expiry-trend", h.GetExpiryTrend)
		dashboard.GET("/alerts", h.GetAlerts)
	}
}

// GetOverview 获取综合概览
// GET /api/v1/dashboard/overview
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	overview, err := h.svc.GetFullOverview()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, overview)
}

// GetCertOverview 获取证书概览
// GET /api/v1/dashboard/cert-overview
func (h *DashboardHandler) GetCertOverview(c *gin.Context) {
	overview, err := h.svc.GetCertOverview()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, overview)
}

// GetDeployOverview 获取部署概览
// GET /api/v1/dashboard/deploy-overview
func (h *DashboardHandler) GetDeployOverview(c *gin.Context) {
	overview, err := h.svc.GetDeployOverview()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, overview)
}

// GetCloudDistribution 获取云资源分布
// GET /api/v1/dashboard/cloud-distribution
func (h *DashboardHandler) GetCloudDistribution(c *gin.Context) {
	distribution, err := h.svc.GetCloudDistribution()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, distribution)
}

// GetExpiryTrend 获取证书到期趋势
// GET /api/v1/dashboard/expiry-trend?days=90
func (h *DashboardHandler) GetExpiryTrend(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "90")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 90
	}

	trend, err := h.svc.GetExpiryTrend(days)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, trend)
}

// GetAlerts 获取告警列表
// GET /api/v1/dashboard/alerts
func (h *DashboardHandler) GetAlerts(c *gin.Context) {
	alerts, err := h.svc.GetAlerts()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, alerts)
}
