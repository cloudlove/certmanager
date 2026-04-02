package repository

import (
	"certmanager-backend/internal/model"

	"gorm.io/gorm"
)

// DomainRepository 域名数据访问层
type DomainRepository struct {
	db *gorm.DB
}

// NewDomainRepository 创建 DomainRepository 实例
func NewDomainRepository(db *gorm.DB) *DomainRepository {
	return &DomainRepository{db: db}
}

// Create 创建域名
func (r *DomainRepository) Create(domain *model.Domain) error {
	return r.db.Create(domain).Error
}

// Update 更新域名
func (r *DomainRepository) Update(domain *model.Domain) error {
	return r.db.Save(domain).Error
}

// Delete 删除域名
func (r *DomainRepository) Delete(id uint) error {
	return r.db.Delete(&model.Domain{}, id).Error
}

// GetByID 根据 ID 获取域名
func (r *DomainRepository) GetByID(id uint) (*model.Domain, error) {
	var domain model.Domain
	if err := r.db.First(&domain, id).Error; err != nil {
		return nil, err
	}
	return &domain, nil
}

// List 分页查询域名列表，支持搜索
func (r *DomainRepository) List(page, pageSize int, search string) ([]model.Domain, int64, error) {
	var domains []model.Domain
	var total int64

	query := r.db.Model(&model.Domain{})
	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&domains).Error; err != nil {
		return nil, 0, err
	}

	return domains, total, nil
}

// GetByIDs 批量查询域名
func (r *DomainRepository) GetByIDs(ids []uint) ([]model.Domain, error) {
	var domains []model.Domain
	if err := r.db.Where("id IN ?", ids).Find(&domains).Error; err != nil {
		return nil, err
	}
	return domains, nil
}

// GetByName 根据名称获取域名
func (r *DomainRepository) GetByName(name string) (*model.Domain, error) {
	var domain model.Domain
	if err := r.db.Where("name = ?", name).First(&domain).Error; err != nil {
		return nil, err
	}
	return &domain, nil
}

// UpdateVerifyStatus 更新域名校验状态
func (r *DomainRepository) UpdateVerifyStatus(id uint, status string) error {
	return r.db.Model(&model.Domain{}).Where("id = ?", id).Update("verify_status", status).Error
}
