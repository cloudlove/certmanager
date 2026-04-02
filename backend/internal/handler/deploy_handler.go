package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"certmanager-backend/internal/model"
	"certmanager-backend/internal/service"
	"certmanager-backend/pkg/response"
)

// CreateDeployTaskRequest 创建部署任务请求
type CreateDeployTaskRequest struct {
	Name          string                 `json:"name" binding:"required"`
	CertificateID uint                   `json:"certificate_id" binding:"required"`
	Targets       []service.DeployTarget `json:"targets" binding:"required,min=1"`
}

// DeployTaskVO 部署任务视图对象
type DeployTaskVO struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	CertificateID uint   `json:"certificate_id"`
	Status        string `json:"status"`
	TotalItems    int    `json:"total_items"`
	SuccessItems  int    `json:"success_items"`
	FailedItems   int    `json:"failed_items"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// DeployTaskItemVO 部署任务项视图对象
type DeployTaskItemVO struct {
	ID           uint   `json:"id"`
	DeployTaskID uint   `json:"deploy_task_id"`
	TargetType   string `json:"target_type"`
	ProviderType string `json:"provider_type"`
	ResourceID   string `json:"resource_id"`
	CredentialID uint   `json:"credential_id"`
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message"`
	SnapshotID   uint   `json:"snapshot_id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// DeployTaskDetailVO 部署任务详情视图对象
type DeployTaskDetailVO struct {
	Task  *DeployTaskVO      `json:"task"`
	Items []DeployTaskItemVO `json:"items"`
}

// DeployHandler 部署任务 HTTP Handler
type DeployHandler struct {
	deployService *service.DeployService
}

// NewDeployHandler 创建 DeployHandler 实例
func NewDeployHandler(deployService *service.DeployService) *DeployHandler {
	return &DeployHandler{
		deployService: deployService,
	}
}

// RegisterRoutes 注册路由
func (h *DeployHandler) RegisterRoutes(r *gin.RouterGroup) {
	// 部署任务路由
	r.POST("/tasks", h.CreateDeployTask)
	r.GET("/tasks", h.ListDeployTasks)
	r.GET("/tasks/:id", h.GetDeployTask)
	r.DELETE("/tasks/:id", h.DeleteDeployTask)
	r.POST("/tasks/:id/execute", h.ExecuteDeployTask)
	r.POST("/tasks/:id/rollback", h.RollbackDeployTask)

	// 部署任务项路由
	r.POST("/task-items/:id/rollback", h.RollbackDeployTaskItem)
}

// toTaskVO 将模型转换为视图对象
func toTaskVO(task *model.DeployTask) *DeployTaskVO {
	return &DeployTaskVO{
		ID:            task.ID,
		Name:          task.Name,
		Type:          task.Type,
		CertificateID: task.CertificateID,
		Status:        task.Status,
		TotalItems:    task.TotalItems,
		SuccessItems:  task.SuccessItems,
		FailedItems:   task.FailedItems,
		CreatedAt:     task.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     task.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// toTaskItemVO 将模型转换为视图对象
func toTaskItemVO(item model.DeployTaskItem) DeployTaskItemVO {
	return DeployTaskItemVO{
		ID:           item.ID,
		DeployTaskID: item.DeployTaskID,
		TargetType:   item.TargetType,
		ProviderType: item.ProviderType,
		ResourceID:   item.ResourceID,
		CredentialID: item.CredentialID,
		Status:       item.Status,
		ErrorMessage: item.ErrorMessage,
		SnapshotID:   item.SnapshotID,
		CreatedAt:    item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// CreateDeployTask 创建部署任务
// POST /api/v1/deploy-tasks
func (h *DeployHandler) CreateDeployTask(c *gin.Context) {
	var req CreateDeployTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	task, err := h.deployService.CreateDeployTask(req.Name, req.CertificateID, req.Targets)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, toTaskVO(task))
}

// ExecuteDeployTask 执行部署任务
// POST /api/v1/deploy-tasks/:id/execute
func (h *DeployHandler) ExecuteDeployTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid task ID")
		return
	}

	if err := h.deployService.ExecuteDeployTask(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "deploy task started"})
}

// RollbackDeployTask 回滚整个任务
// POST /api/v1/deploy-tasks/:id/rollback
func (h *DeployHandler) RollbackDeployTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid task ID")
		return
	}

	if err := h.deployService.RollbackDeployTask(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "rollback task started"})
}

// RollbackDeployTaskItem 回滚单个子任务
// POST /api/v1/deploy-task-items/:id/rollback
func (h *DeployHandler) RollbackDeployTaskItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid item ID")
		return
	}

	if err := h.deployService.RollbackDeployTaskItem(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "rollback item started"})
}

// GetDeployTask 获取任务详情（含子任务）
// GET /api/v1/deploy-tasks/:id
func (h *DeployHandler) GetDeployTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid task ID")
		return
	}

	detail, err := h.deployService.GetDeployTask(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	// 转换为视图对象
	itemVOs := make([]DeployTaskItemVO, 0, len(detail.Items))
	for _, item := range detail.Items {
		itemVOs = append(itemVOs, toTaskItemVO(item))
	}

	vo := DeployTaskDetailVO{
		Task:  toTaskVO(detail.Task),
		Items: itemVOs,
	}

	response.Success(c, vo)
}

// ListDeployTasks 分页查询部署任务
// GET /api/v1/deploy-tasks?page=1&pageSize=10&status=pending
func (h *DeployHandler) ListDeployTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	tasks, total, err := h.deployService.ListDeployTasks(page, pageSize, status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 转换为视图对象
	vos := make([]*DeployTaskVO, 0, len(tasks))
	for i := range tasks {
		vos = append(vos, toTaskVO(&tasks[i]))
	}

	response.Success(c, gin.H{
		"list":  vos,
		"total": total,
	})
}

// DeleteDeployTask 删除部署任务
// DELETE /api/v1/deploy-tasks/:id
func (h *DeployHandler) DeleteDeployTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid task ID")
		return
	}

	if err := h.deployService.DeleteDeployTask(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "deploy task deleted"})
}

// ==================== WebSocket Handler ====================

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境应该限制
	},
}

// DeployWebSocketHandler WebSocket 部署进度处理器
type DeployWebSocketHandler struct {
	deployService *service.DeployService
}

// NewDeployWebSocketHandler 创建 DeployWebSocketHandler 实例
func NewDeployWebSocketHandler(deployService *service.DeployService) *DeployWebSocketHandler {
	return &DeployWebSocketHandler{
		deployService: deployService,
	}
}

// RegisterWebSocketRoutes 注册 WebSocket 路由
func (h *DeployWebSocketHandler) RegisterWebSocketRoutes(r *gin.Engine) {
	r.GET("/ws/deploy/:id", h.HandleDeployProgress)
}

// HandleDeployProgress WebSocket 处理部署进度
// GET /ws/deploy/:id
func (h *DeployWebSocketHandler) HandleDeployProgress(c *gin.Context) {
	idStr := c.Param("id")
	taskID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	// 升级 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 注册进度通道
	progressCh := h.deployService.RegisterProgressChannel(uint(taskID))
	defer h.deployService.UnregisterProgressChannel(uint(taskID))

	// 设置写入超时
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// 发送初始消息
	if err := conn.WriteJSON(gin.H{
		"type":    "connected",
		"task_id": taskID,
		"message": "connected to deploy progress stream",
	}); err != nil {
		return
	}

	// 监听进度更新
	for progress := range progressCh {
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := conn.WriteJSON(gin.H{
			"type":      "progress",
			"task_id":   progress.TaskID,
			"item_id":   progress.ItemID,
			"status":    progress.Status,
			"message":   progress.Message,
			"timestamp": progress.Timestamp,
		}); err != nil {
			// 写入失败，客户端可能已断开
			return
		}
	}

	// 通道关闭，发送完成消息
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	conn.WriteJSON(gin.H{
		"type":    "completed",
		"task_id": taskID,
		"message": "deploy progress stream closed",
	})
}
