package repository

import (
	"certmanager-backend/internal/model"

	"gorm.io/gorm"
)

// DeployRepository 部署任务数据访问层
type DeployRepository struct {
	db *gorm.DB
}

// NewDeployRepository 创建 DeployRepository 实例
func NewDeployRepository(db *gorm.DB) *DeployRepository {
	return &DeployRepository{db: db}
}

// ==================== Task 相关操作 ====================

// CreateTask 创建部署任务
func (r *DeployRepository) CreateTask(task *model.DeployTask) error {
	return r.db.Create(task).Error
}

// UpdateTask 更新部署任务
func (r *DeployRepository) UpdateTask(task *model.DeployTask) error {
	return r.db.Save(task).Error
}

// GetTaskByID 根据 ID 获取部署任务
func (r *DeployRepository) GetTaskByID(id uint) (*model.DeployTask, error) {
	var task model.DeployTask
	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// ListTasks 分页查询部署任务列表，可选状态筛选
func (r *DeployRepository) ListTasks(page, pageSize int, status string) ([]model.DeployTask, int64, error) {
	var tasks []model.DeployTask
	var total int64

	query := r.db.Model(&model.DeployTask{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// DeleteTask 删除部署任务
func (r *DeployRepository) DeleteTask(id uint) error {
	return r.db.Delete(&model.DeployTask{}, id).Error
}

// ==================== TaskItem 相关操作 ====================

// CreateTaskItem 创建部署任务项
func (r *DeployRepository) CreateTaskItem(item *model.DeployTaskItem) error {
	return r.db.Create(item).Error
}

// UpdateTaskItem 更新部署任务项
func (r *DeployRepository) UpdateTaskItem(item *model.DeployTaskItem) error {
	return r.db.Save(item).Error
}

// GetTaskItemByID 根据 ID 获取部署任务项
func (r *DeployRepository) GetTaskItemByID(id uint) (*model.DeployTaskItem, error) {
	var item model.DeployTaskItem
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// GetTaskItemsByTaskID 根据任务 ID 获取所有任务项
func (r *DeployRepository) GetTaskItemsByTaskID(taskID uint) ([]model.DeployTaskItem, error) {
	var items []model.DeployTaskItem
	if err := r.db.Where("deploy_task_id = ?", taskID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// CountTaskItemsByStatus 统计任务项各状态数量
func (r *DeployRepository) CountTaskItemsByStatus(taskID uint) (map[string]int64, error) {
	result := make(map[string]int64)

	var counts []struct {
		Status string
		Count  int64
	}

	if err := r.db.Model(&model.DeployTaskItem{}).
		Where("deploy_task_id = ?", taskID).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&counts).Error; err != nil {
		return nil, err
	}

	for _, c := range counts {
		result[c.Status] = c.Count
	}

	return result, nil
}

// ==================== Snapshot 相关操作 ====================

// CreateSnapshot 创建部署快照
func (r *DeployRepository) CreateSnapshot(snapshot *model.DeploySnapshot) error {
	return r.db.Create(snapshot).Error
}

// GetSnapshotByItemID 根据任务项 ID 获取快照
func (r *DeployRepository) GetSnapshotByItemID(itemID uint) (*model.DeploySnapshot, error) {
	var snapshot model.DeploySnapshot
	if err := r.db.Where("deploy_task_item_id = ?", itemID).First(&snapshot).Error; err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// UpdateTaskStatusAndCounts 更新任务状态和计数
func (r *DeployRepository) UpdateTaskStatusAndCounts(taskID uint, status string, successCount, failedCount int) error {
	return r.db.Model(&model.DeployTask{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status":        status,
			"success_items": successCount,
			"failed_items":  failedCount,
		}).Error
}
