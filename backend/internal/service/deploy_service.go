package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"certmanager-backend/internal/model"
	"certmanager-backend/internal/provider/cloud"
	"certmanager-backend/internal/repository"
	"certmanager-backend/pkg/certutil"
	"certmanager-backend/pkg/crypto"
)

// DeployTarget 部署目标
type DeployTarget struct {
	ProviderType string `json:"provider_type"`
	TargetType   string `json:"target_type"`
	ResourceID   string `json:"resource_id"`
	CredentialID uint   `json:"credential_id"`
}

// DeployTaskDetail 部署任务详情（包含子任务）
type DeployTaskDetail struct {
	Task  *model.DeployTask      `json:"task"`
	Items []model.DeployTaskItem `json:"items"`
}

// DeployProgress 部署进度更新
type DeployProgress struct {
	TaskID    uint   `json:"task_id"`
	ItemID    uint   `json:"item_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// DeployService 部署业务逻辑层
type DeployService struct {
	deployRepo     *repository.DeployRepository
	credentialRepo *repository.CredentialRepository
	certRepo       *repository.CertificateRepository
	aesKey         string
	// WebSocket 通知通道
	progressChannels map[uint]chan DeployProgress
	channelsMu       sync.RWMutex
	// Goroutine 池
	workerPool chan struct{}
}

// NewDeployService 创建 DeployService 实例
func NewDeployService(
	deployRepo *repository.DeployRepository,
	credentialRepo *repository.CredentialRepository,
	certRepo *repository.CertificateRepository,
	aesKey string,
) *DeployService {
	return &DeployService{
		deployRepo:       deployRepo,
		credentialRepo:   credentialRepo,
		certRepo:         certRepo,
		aesKey:           aesKey,
		progressChannels: make(map[uint]chan DeployProgress),
		workerPool:       make(chan struct{}, 10), // 限制并发数为 10
	}
}

// RegisterProgressChannel 注册任务进度通知通道
func (s *DeployService) RegisterProgressChannel(taskID uint) chan DeployProgress {
	ch := make(chan DeployProgress, 100)
	s.channelsMu.Lock()
	s.progressChannels[taskID] = ch
	s.channelsMu.Unlock()
	return ch
}

// UnregisterProgressChannel 注销任务进度通知通道
func (s *DeployService) UnregisterProgressChannel(taskID uint) {
	s.channelsMu.Lock()
	if ch, ok := s.progressChannels[taskID]; ok {
		close(ch)
		delete(s.progressChannels, taskID)
	}
	s.channelsMu.Unlock()
}

// notifyProgress 发送进度通知
func (s *DeployService) notifyProgress(progress DeployProgress) {
	s.channelsMu.RLock()
	ch, ok := s.progressChannels[progress.TaskID]
	s.channelsMu.RUnlock()
	if ok {
		select {
		case ch <- progress:
		default:
			// 通道已满，丢弃消息
		}
	}
}

// CreateDeployTask 创建部署任务，拆分为子任务
func (s *DeployService) CreateDeployTask(name string, certID uint, targets []DeployTarget) (*model.DeployTask, error) {
	if name == "" {
		return nil, errors.New("task name is required")
	}
	if certID == 0 {
		return nil, errors.New("certificate ID is required")
	}
	if len(targets) == 0 {
		return nil, errors.New("at least one target is required")
	}

	// 验证证书是否存在
	_, err := s.certRepo.GetByID(certID)
	if err != nil {
		return nil, fmt.Errorf("certificate not found: %w", err)
	}

	// 验证每个目标的凭证和资源类型
	for i, target := range targets {
		if target.ProviderType == "" || target.ResourceID == "" || target.TargetType == "" {
			return nil, fmt.Errorf("target %d: provider_type, resource_id and target_type are required", i)
		}
		if !cloud.IsValidResourceType(target.ProviderType, target.TargetType) {
			return nil, fmt.Errorf("target %d: invalid resource type %s for provider %s", i, target.TargetType, target.ProviderType)
		}
		_, err := s.credentialRepo.GetByID(target.CredentialID)
		if err != nil {
			return nil, fmt.Errorf("target %d: credential not found: %w", i, err)
		}
	}

	// 创建主任务
	task := &model.DeployTask{
		Name:          name,
		Type:          "cloud",
		CertificateID: certID,
		Status:        "pending",
		TotalItems:    len(targets),
		SuccessItems:  0,
		FailedItems:   0,
	}

	if err := s.deployRepo.CreateTask(task); err != nil {
		return nil, fmt.Errorf("failed to create deploy task: %w", err)
	}

	// 创建子任务
	for _, target := range targets {
		item := &model.DeployTaskItem{
			DeployTaskID: task.ID,
			TargetType:   target.TargetType,
			ProviderType: target.ProviderType,
			ResourceID:   target.ResourceID,
			CredentialID: target.CredentialID,
			Status:       "pending",
		}
		if err := s.deployRepo.CreateTaskItem(item); err != nil {
			return nil, fmt.Errorf("failed to create task item: %w", err)
		}
	}

	return task, nil
}

// ExecuteDeployTask 异步执行部署
func (s *DeployService) ExecuteDeployTask(taskID uint) error {
	// 获取任务
	task, err := s.deployRepo.GetTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	if task.Status != "pending" {
		return fmt.Errorf("task status is %s, cannot execute", task.Status)
	}

	// 获取证书
	cert, err := s.certRepo.GetByID(task.CertificateID)
	if err != nil {
		return fmt.Errorf("certificate not found: %w", err)
	}

	// 解密私钥
	privateKey, err := crypto.Decrypt(cert.PrivateKeyEncrypted, s.aesKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt private key: %w", err)
	}

	// 获取所有子任务
	items, err := s.deployRepo.GetTaskItemsByTaskID(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task items: %w", err)
	}

	// 更新任务状态为 deploying
	task.Status = "deploying"
	if err := s.deployRepo.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// 异步执行部署
	go s.executeDeployAsync(task, items, cert.CertPEM, cert.ChainPEM, privateKey)

	return nil
}

// executeDeployAsync 异步执行部署
func (s *DeployService) executeDeployAsync(task *model.DeployTask, items []model.DeployTaskItem, certPEM, chainPEM, privateKey string) {
	var wg sync.WaitGroup
	successCount := 0
	failedCount := 0
	var mu sync.Mutex

	for i := range items {
		wg.Add(1)
		s.workerPool <- struct{}{} // 获取 worker

		go func(item *model.DeployTaskItem) {
			defer wg.Done()
			defer func() { <-s.workerPool }() // 释放 worker

			s.executeDeployItem(task.ID, item, certPEM, chainPEM, privateKey)

			mu.Lock()
			if item.Status == "success" {
				successCount++
			} else if item.Status == "failed" {
				failedCount++
			}
			mu.Unlock()
		}(&items[i])
	}

	wg.Wait()

	// 更新任务最终状态
	var finalStatus string
	if successCount == len(items) {
		finalStatus = "success"
	} else if successCount > 0 {
		finalStatus = "partial_success"
	} else {
		finalStatus = "failed"
	}

	s.deployRepo.UpdateTaskStatusAndCounts(task.ID, finalStatus, successCount, failedCount)

	// 发送最终进度通知
	s.notifyProgress(DeployProgress{
		TaskID:    task.ID,
		Status:    finalStatus,
		Message:   fmt.Sprintf("Deploy completed: %d success, %d failed", successCount, failedCount),
		Timestamp: time.Now().Unix(),
	})
}

// executeDeployItem 执行单个部署项
func (s *DeployService) executeDeployItem(taskID uint, item *model.DeployTaskItem, certPEM, chainPEM, privateKey string) {
	ctx := context.Background()

	// 更新状态为 deploying
	item.Status = "deploying"
	s.deployRepo.UpdateTaskItem(item)

	s.notifyProgress(DeployProgress{
		TaskID:    taskID,
		ItemID:    item.ID,
		Status:    "deploying",
		Message:   fmt.Sprintf("Deploying to %s %s", item.ProviderType, item.ResourceID),
		Timestamp: time.Now().Unix(),
	})

	// 获取凭证
	credential, err := s.credentialRepo.GetByID(item.CredentialID)
	if err != nil {
		item.Status = "failed"
		item.ErrorMessage = fmt.Sprintf("failed to get credential: %v", err)
		s.deployRepo.UpdateTaskItem(item)
		s.notifyProgress(DeployProgress{
			TaskID:    taskID,
			ItemID:    item.ID,
			Status:    "failed",
			Message:   item.ErrorMessage,
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 解密凭证
	accessKey, err := crypto.Decrypt(credential.AccessKeyEncrypted, s.aesKey)
	if err != nil {
		item.Status = "failed"
		item.ErrorMessage = fmt.Sprintf("failed to decrypt access key: %v", err)
		s.deployRepo.UpdateTaskItem(item)
		s.notifyProgress(DeployProgress{
			TaskID:    taskID,
			ItemID:    item.ID,
			Status:    "failed",
			Message:   item.ErrorMessage,
			Timestamp: time.Now().Unix(),
		})
		return
	}

	secretKey, err := crypto.Decrypt(credential.SecretKeyEncrypted, s.aesKey)
	if err != nil {
		item.Status = "failed"
		item.ErrorMessage = fmt.Sprintf("failed to decrypt secret key: %v", err)
		s.deployRepo.UpdateTaskItem(item)
		s.notifyProgress(DeployProgress{
			TaskID:    taskID,
			ItemID:    item.ID,
			Status:    "failed",
			Message:   item.ErrorMessage,
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 创建 Provider
	provider, err := cloud.NewCloudDeployProvider(item.ProviderType, accessKey, secretKey)
	if err != nil {
		item.Status = "failed"
		item.ErrorMessage = fmt.Sprintf("failed to create provider: %v", err)
		s.deployRepo.UpdateTaskItem(item)
		s.notifyProgress(DeployProgress{
			TaskID:    taskID,
			ItemID:    item.ID,
			Status:    "failed",
			Message:   item.ErrorMessage,
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 获取旧证书快照
	oldCertInfo, err := provider.GetCurrentCert(ctx, item.ResourceID, item.TargetType)
	if err != nil {
		// 快照失败不阻塞部署，记录警告
		oldCertInfo = ""
	}

	// 保存快照
	snapshot := &model.DeploySnapshot{
		DeployTaskItemID: item.ID,
		OldCertInfo:      oldCertInfo,
	}
	if err := s.deployRepo.CreateSnapshot(snapshot); err != nil {
		// 快照保存失败不阻塞部署
		fmt.Printf("Failed to create snapshot: %v\n", err)
	}
	item.SnapshotID = snapshot.ID
	s.deployRepo.UpdateTaskItem(item)

	// 执行部署
	deployReq := cloud.DeployRequest{
		CertPEM:       certPEM,
		ChainPEM:      chainPEM,
		PrivateKeyPEM: privateKey,
		ResourceID:    item.ResourceID,
		ResourceType:  item.TargetType,
	}

	resp, err := provider.DeployCert(ctx, deployReq)
	if err != nil {
		item.Status = "failed"
		item.ErrorMessage = fmt.Sprintf("deploy failed: %v", err)
		s.deployRepo.UpdateTaskItem(item)
		s.notifyProgress(DeployProgress{
			TaskID:    taskID,
			ItemID:    item.ID,
			Status:    "failed",
			Message:   item.ErrorMessage,
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 部署后校验 - 尝试解析证书
	if _, err := certutil.ParseCertificate(certPEM); err != nil {
		item.Status = "failed"
		item.ErrorMessage = fmt.Sprintf("certificate validation failed: %v", err)
		s.deployRepo.UpdateTaskItem(item)
		s.notifyProgress(DeployProgress{
			TaskID:    taskID,
			ItemID:    item.ID,
			Status:    "failed",
			Message:   item.ErrorMessage,
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 更新状态为成功
	item.Status = "success"
	item.ErrorMessage = resp.Message
	s.deployRepo.UpdateTaskItem(item)

	s.notifyProgress(DeployProgress{
		TaskID:    taskID,
		ItemID:    item.ID,
		Status:    "success",
		Message:   fmt.Sprintf("Deployed to %s %s successfully", item.ProviderType, item.ResourceID),
		Timestamp: time.Now().Unix(),
	})
}

// RollbackDeployTask 回滚整个任务
func (s *DeployService) RollbackDeployTask(taskID uint) error {
	// 验证任务存在
	_, err := s.deployRepo.GetTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// 获取所有成功的子任务
	items, err := s.deployRepo.GetTaskItemsByTaskID(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task items: %w", err)
	}

	var rollbackItems []model.DeployTaskItem
	for _, item := range items {
		if item.Status == "success" {
			rollbackItems = append(rollbackItems, item)
		}
	}

	if len(rollbackItems) == 0 {
		return errors.New("no successful items to rollback")
	}

	// 异步执行回滚
	go s.executeRollbackAsync(taskID, rollbackItems)

	return nil
}

// executeRollbackAsync 异步执行回滚
func (s *DeployService) executeRollbackAsync(taskID uint, items []model.DeployTaskItem) {
	var wg sync.WaitGroup

	for i := range items {
		wg.Add(1)
		s.workerPool <- struct{}{}

		go func(item *model.DeployTaskItem) {
			defer wg.Done()
			defer func() { <-s.workerPool }()

			s.executeRollbackItem(taskID, item)
		}(&items[i])
	}

	wg.Wait()

	s.notifyProgress(DeployProgress{
		TaskID:    taskID,
		Status:    "rollback_completed",
		Message:   "Rollback completed",
		Timestamp: time.Now().Unix(),
	})
}

// executeRollbackItem 执行单个子任务回滚
func (s *DeployService) executeRollbackItem(taskID uint, item *model.DeployTaskItem) {
	ctx := context.Background()

	s.notifyProgress(DeployProgress{
		TaskID:    taskID,
		ItemID:    item.ID,
		Status:    "rolling_back",
		Message:   fmt.Sprintf("Rolling back %s %s", item.ProviderType, item.ResourceID),
		Timestamp: time.Now().Unix(),
	})

	// 获取快照
	snapshot, err := s.deployRepo.GetSnapshotByItemID(item.ID)
	if err != nil {
		item.Status = "failed"
		item.ErrorMessage = fmt.Sprintf("snapshot not found: %v", err)
		s.deployRepo.UpdateTaskItem(item)
		s.notifyProgress(DeployProgress{
			TaskID:    taskID,
			ItemID:    item.ID,
			Status:    "rollback_failed",
			Message:   item.ErrorMessage,
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 获取凭证
	credential, err := s.credentialRepo.GetByID(item.CredentialID)
	if err != nil {
		item.Status = "failed"
		item.ErrorMessage = fmt.Sprintf("failed to get credential: %v", err)
		s.deployRepo.UpdateTaskItem(item)
		s.notifyProgress(DeployProgress{
			TaskID:    taskID,
			ItemID:    item.ID,
			Status:    "rollback_failed",
			Message:   item.ErrorMessage,
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 解密凭证
	accessKey, _ := crypto.Decrypt(credential.AccessKeyEncrypted, s.aesKey)
	secretKey, _ := crypto.Decrypt(credential.SecretKeyEncrypted, s.aesKey)

	// 创建 Provider
	provider, err := cloud.NewCloudDeployProvider(item.ProviderType, accessKey, secretKey)
	if err != nil {
		item.Status = "failed"
		item.ErrorMessage = fmt.Sprintf("failed to create provider: %v", err)
		s.deployRepo.UpdateTaskItem(item)
		s.notifyProgress(DeployProgress{
			TaskID:    taskID,
			ItemID:    item.ID,
			Status:    "rollback_failed",
			Message:   item.ErrorMessage,
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 执行回滚
	rollbackReq := cloud.RollbackRequest{
		ResourceID:   item.ResourceID,
		ResourceType: item.TargetType,
		OldCertInfo:  snapshot.OldCertInfo,
	}

	if err := provider.Rollback(ctx, rollbackReq); err != nil {
		item.Status = "failed"
		item.ErrorMessage = fmt.Sprintf("rollback failed: %v", err)
		s.deployRepo.UpdateTaskItem(item)
		s.notifyProgress(DeployProgress{
			TaskID:    taskID,
			ItemID:    item.ID,
			Status:    "rollback_failed",
			Message:   item.ErrorMessage,
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 更新状态为已回滚
	item.Status = "rolled_back"
	item.ErrorMessage = ""
	s.deployRepo.UpdateTaskItem(item)

	s.notifyProgress(DeployProgress{
		TaskID:    taskID,
		ItemID:    item.ID,
		Status:    "rolled_back",
		Message:   fmt.Sprintf("Rolled back %s %s successfully", item.ProviderType, item.ResourceID),
		Timestamp: time.Now().Unix(),
	})
}

// RollbackDeployTaskItem 回滚单个子任务
func (s *DeployService) RollbackDeployTaskItem(itemID uint) error {
	item, err := s.deployRepo.GetTaskItemByID(itemID)
	if err != nil {
		return fmt.Errorf("task item not found: %w", err)
	}

	if item.Status != "success" {
		return fmt.Errorf("item status is %s, cannot rollback", item.Status)
	}

	go s.executeRollbackItem(item.DeployTaskID, item)

	return nil
}

// GetDeployTask 获取任务详情（含子任务列表）
func (s *DeployService) GetDeployTask(id uint) (*DeployTaskDetail, error) {
	task, err := s.deployRepo.GetTaskByID(id)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	items, err := s.deployRepo.GetTaskItemsByTaskID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task items: %w", err)
	}

	return &DeployTaskDetail{
		Task:  task,
		Items: items,
	}, nil
}

// ListDeployTasks 分页查询部署任务
func (s *DeployService) ListDeployTasks(page, pageSize int, status string) ([]model.DeployTask, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	return s.deployRepo.ListTasks(page, pageSize, status)
}

// DeleteDeployTask 删除部署任务
func (s *DeployService) DeleteDeployTask(id uint) error {
	// 获取任务
	task, err := s.deployRepo.GetTaskByID(id)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// 如果任务正在执行中，不能删除
	if task.Status == "deploying" {
		return errors.New("cannot delete task while deploying")
	}

	return s.deployRepo.DeleteTask(id)
}
