package handler

import (
	"strconv"

	"certmanager-backend/internal/service"
	"certmanager-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// NginxHandler Nginx HTTP Handler
type NginxHandler struct {
	svc *service.NginxService
}

// NewNginxHandler 创建 NginxHandler 实例
func NewNginxHandler(svc *service.NginxService) *NginxHandler {
	return &NginxHandler{svc: svc}
}

// RegisterRoutes 注册 Nginx 路由
func (h *NginxHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// 集群路由
	rg.POST("/clusters", h.CreateCluster)
	rg.GET("/clusters", h.ListClusters)
	rg.GET("/clusters/:id", h.GetCluster)
	rg.DELETE("/clusters/:id", h.DeleteCluster)
	rg.POST("/clusters/:id/nodes", h.AddNode)
	rg.POST("/clusters/:id/deploy", h.DeployToCluster)

	// 节点路由
	rg.DELETE("/nodes/:id", h.RemoveNode)

	// 心跳路由
	rg.POST("/heartbeat", h.ReceiveHeartbeat)
}

// ==================== 请求/响应结构体 ====================

// createClusterReq 创建集群请求
type createClusterReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// addNodeReq 添加节点请求
type addNodeReq struct {
	IP   string `json:"ip" binding:"required"`
	Port string `json:"port" binding:"required"`
}

// deployReq 部署证书请求
type deployReq struct {
	CertificateID uint `json:"certificate_id" binding:"required"`
}

// heartbeatReq 心跳请求
type heartbeatReq struct {
	IP     string `json:"ip" binding:"required"`
	Port   string `json:"port" binding:"required"`
	Status string `json:"status"`
}

// ==================== Handler 方法 ====================

// CreateCluster 创建集群
// POST /api/v1/nginx/clusters
func (h *NginxHandler) CreateCluster(c *gin.Context) {
	var req createClusterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	vo, err := h.svc.CreateCluster(req.Name, req.Description)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, vo)
}

// ListClusters 获取集群列表
// GET /api/v1/nginx/clusters?page=1&pageSize=10
func (h *NginxHandler) ListClusters(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	vos, total, err := h.svc.ListClusters(page, pageSize)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.SuccessWithPage(c, vos, total, page, pageSize)
}

// GetCluster 获取集群详情
// GET /api/v1/nginx/clusters/:id
func (h *NginxHandler) GetCluster(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	vo, err := h.svc.GetCluster(id)
	if err != nil {
		response.Error(c, 404, err.Error())
		return
	}

	response.Success(c, vo)
}

// DeleteCluster 删除集群
// DELETE /api/v1/nginx/clusters/:id
func (h *NginxHandler) DeleteCluster(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	if err := h.svc.DeleteCluster(id); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, nil)
}

// AddNode 添加节点
// POST /api/v1/nginx/clusters/:id/nodes
func (h *NginxHandler) AddNode(c *gin.Context) {
	clusterID, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid cluster id")
		return
	}

	var req addNodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	vo, err := h.svc.AddNode(clusterID, req.IP, req.Port)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, vo)
}

// RemoveNode 移除节点
// DELETE /api/v1/nginx/nodes/:id
func (h *NginxHandler) RemoveNode(c *gin.Context) {
	nodeID, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid node id")
		return
	}

	if err := h.svc.RemoveNode(nodeID); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeployToCluster 部署证书到集群
// POST /api/v1/nginx/clusters/:id/deploy
func (h *NginxHandler) DeployToCluster(c *gin.Context) {
	clusterID, err := parseID(c)
	if err != nil {
		response.Error(c, 400, "invalid cluster id")
		return
	}

	var req deployReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	results, err := h.svc.DeployToCluster(clusterID, req.CertificateID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, results)
}

// ReceiveHeartbeat 接收节点心跳
// POST /api/v1/nginx/heartbeat
func (h *NginxHandler) ReceiveHeartbeat(c *gin.Context) {
	var req heartbeatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request: "+err.Error())
		return
	}

	if req.Status == "" {
		req.Status = "online"
	}

	if err := h.svc.ReceiveHeartbeat(req.IP, req.Port, req.Status); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, nil)
}
