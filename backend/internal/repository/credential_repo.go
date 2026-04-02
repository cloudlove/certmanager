package repository

import (
	"certmanager-backend/internal/model"

	"gorm.io/gorm"
)

// CredentialRepository 云凭证数据访问层
type CredentialRepository struct {
	db *gorm.DB
}

// NewCredentialRepository 创建 CredentialRepository 实例
func NewCredentialRepository(db *gorm.DB) *CredentialRepository {
	return &CredentialRepository{db: db}
}

// Create 创建凭证
func (r *CredentialRepository) Create(credential *model.CloudCredential) error {
	return r.db.Create(credential).Error
}

// Update 更新凭证
func (r *CredentialRepository) Update(credential *model.CloudCredential) error {
	return r.db.Save(credential).Error
}

// Delete 删除凭证
func (r *CredentialRepository) Delete(id uint) error {
	return r.db.Delete(&model.CloudCredential{}, id).Error
}

// GetByID 根据 ID 获取凭证
func (r *CredentialRepository) GetByID(id uint) (*model.CloudCredential, error) {
	var credential model.CloudCredential
	if err := r.db.First(&credential, id).Error; err != nil {
		return nil, err
	}
	return &credential, nil
}

// List 分页查询凭证列表，可选 providerType 筛选
func (r *CredentialRepository) List(page, pageSize int, providerType string) ([]model.CloudCredential, int64, error) {
	var credentials []model.CloudCredential
	var total int64

	query := r.db.Model(&model.CloudCredential{})
	if providerType != "" {
		query = query.Where("provider_type = ?", providerType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&credentials).Error; err != nil {
		return nil, 0, err
	}

	return credentials, total, nil
}
