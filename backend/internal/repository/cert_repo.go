package repository

import (
	"certmanager-backend/internal/model"

	"gorm.io/gorm"
)

// CertificateRepository 证书数据访问层
type CertificateRepository struct {
	db *gorm.DB
}

// NewCertificateRepository 创建 CertificateRepository 实例
func NewCertificateRepository(db *gorm.DB) *CertificateRepository {
	return &CertificateRepository{db: db}
}

// Create 创建证书记录
func (r *CertificateRepository) Create(cert *model.Certificate) error {
	return r.db.Create(cert).Error
}

// Update 更新证书记录
func (r *CertificateRepository) Update(cert *model.Certificate) error {
	return r.db.Save(cert).Error
}

// Delete 删除证书记录
func (r *CertificateRepository) Delete(id uint) error {
	return r.db.Delete(&model.Certificate{}, id).Error
}

// GetByID 根据 ID 获取证书
func (r *CertificateRepository) GetByID(id uint) (*model.Certificate, error) {
	var cert model.Certificate
	if err := r.db.First(&cert, id).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

// GetByIDs 批量查询证书
func (r *CertificateRepository) GetByIDs(ids []uint) ([]model.Certificate, error) {
	var certs []model.Certificate
	if err := r.db.Where("id IN ?", ids).Find(&certs).Error; err != nil {
		return nil, err
	}
	return certs, nil
}

// GetByDomain 根据域名获取证书
func (r *CertificateRepository) GetByDomain(domain string) (*model.Certificate, error) {
	var cert model.Certificate
	if err := r.db.Where("domain = ?", domain).First(&cert).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

// GetCertPEMByID 获取证书 PEM 内容
func (r *CertificateRepository) GetCertPEMByID(id uint) (string, error) {
	var cert model.Certificate
	if err := r.db.Select("cert_pem").First(&cert, id).Error; err != nil {
		return "", err
	}
	return cert.CertPEM, nil
}

// List 分页查询证书列表，支持状态筛选和搜索
func (r *CertificateRepository) List(page, pageSize int, status, search, sortBy string) ([]model.Certificate, int64, error) {
	var certs []model.Certificate
	var total int64

	query := r.db.Model(&model.Certificate{})

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 搜索（域名或颁发者）
	if search != "" {
		query = query.Where("domain LIKE ? OR issuer LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderBy := "id DESC"
	if sortBy != "" {
		switch sortBy {
		case "expire_asc":
			orderBy = "expire_at ASC"
		case "expire_desc":
			orderBy = "expire_at DESC"
		case "created_asc":
			orderBy = "created_at ASC"
		case "created_desc":
			orderBy = "created_at DESC"
		default:
			orderBy = "id DESC"
		}
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order(orderBy).Find(&certs).Error; err != nil {
		return nil, 0, err
	}

	return certs, total, nil
}

// ListAll 获取所有证书列表
func (r *CertificateRepository) ListAll() ([]model.Certificate, error) {
	var certs []model.Certificate
	if err := r.db.Find(&certs).Error; err != nil {
		return nil, err
	}
	return certs, nil
}

// ListByStatuses 根据状态列表查询证书
func (r *CertificateRepository) ListByStatuses(statuses []string) ([]model.Certificate, error) {
	var certs []model.Certificate
	if err := r.db.Where("status IN ?", statuses).Find(&certs).Error; err != nil {
		return nil, err
	}
	return certs, nil
}

// CountByStatus 统计各状态证书数量
func (r *CertificateRepository) CountByStatus() (map[string]int64, error) {
	result := make(map[string]int64)

	var counts []struct {
		Status string
		Count  int64
	}

	if err := r.db.Model(&model.Certificate{}).
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

// CertRepository 证书数据访问层（辅助查询，兼容旧代码）
type CertRepository struct {
	db *gorm.DB
}

// NewCertRepository 创建 CertRepository 实例
func NewCertRepository(db *gorm.DB) *CertRepository {
	return &CertRepository{db: db}
}

// GetByID 根据 ID 获取证书
func (r *CertRepository) GetByID(id uint) (*model.Certificate, error) {
	var cert model.Certificate
	if err := r.db.First(&cert, id).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

// GetByIDs 批量查询证书
func (r *CertRepository) GetByIDs(ids []uint) ([]model.Certificate, error) {
	var certs []model.Certificate
	if err := r.db.Where("id IN ?", ids).Find(&certs).Error; err != nil {
		return nil, err
	}
	return certs, nil
}

// GetByDomain 根据域名获取证书
func (r *CertRepository) GetByDomain(domain string) (*model.Certificate, error) {
	var cert model.Certificate
	if err := r.db.Where("domain = ?", domain).First(&cert).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

// GetCertPEMByID 获取证书 PEM 内容
func (r *CertRepository) GetCertPEMByID(id uint) (string, error) {
	var cert model.Certificate
	if err := r.db.Select("cert_pem").First(&cert, id).Error; err != nil {
		return "", err
	}
	return cert.CertPEM, nil
}

// ListAll 获取所有证书列表
func (r *CertRepository) ListAll() ([]model.Certificate, error) {
	var certs []model.Certificate
	if err := r.db.Find(&certs).Error; err != nil {
		return nil, err
	}
	return certs, nil
}
