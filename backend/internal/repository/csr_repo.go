package repository

import (
	"certmanager-backend/internal/model"

	"gorm.io/gorm"
)

// CSRRepository CSR 数据访问层
type CSRRepository struct {
	db *gorm.DB
}

// NewCSRRepository 创建 CSRRepository 实例
func NewCSRRepository(db *gorm.DB) *CSRRepository {
	return &CSRRepository{db: db}
}

// Create 创建 CSR 记录
func (r *CSRRepository) Create(csr *model.CSRRecord) error {
	return r.db.Create(csr).Error
}

// Delete 删除 CSR 记录
func (r *CSRRepository) Delete(id uint) error {
	return r.db.Delete(&model.CSRRecord{}, id).Error
}

// GetByID 根据 ID 获取 CSR 记录
func (r *CSRRepository) GetByID(id uint) (*model.CSRRecord, error) {
	var csr model.CSRRecord
	if err := r.db.First(&csr, id).Error; err != nil {
		return nil, err
	}
	return &csr, nil
}

// List 分页查询 CSR 列表，支持按 CommonName 模糊搜索
func (r *CSRRepository) List(page, pageSize int, search string) ([]model.CSRRecord, int64, error) {
	var csrs []model.CSRRecord
	var total int64

	query := r.db.Model(&model.CSRRecord{})
	if search != "" {
		query = query.Where("common_name LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&csrs).Error; err != nil {
		return nil, 0, err
	}

	return csrs, total, nil
}
