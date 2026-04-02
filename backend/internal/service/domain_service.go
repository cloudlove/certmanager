package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"certmanager-backend/internal/model"
	"certmanager-backend/internal/repository"
	"certmanager-backend/pkg/certutil"
)

// VerifyStatus 域名校验状态常量
const (
	VerifyStatusNormal    = "normal"    // 一致
	VerifyStatusMismatch  = "mismatch"  // 不匹配
	VerifyStatusExpired   = "expired"   // 已过期
	VerifyStatusError     = "error"     // 校验失败
	VerifyStatusUnchecked = "unchecked" // 未校验
)

// DomainVO 域名视图对象
type DomainVO struct {
	ID            uint       `json:"id"`
	Name          string     `json:"name"`
	CertificateID uint       `json:"certificate_id"`
	VerifyStatus  string     `json:"verify_status"`
	LastCheckAt   *time.Time `json:"last_check_at"`
	CreatedAt     string     `json:"created_at"`
	UpdatedAt     string     `json:"updated_at"`
}

// VerifyResult 校验结果
type VerifyResult struct {
	DomainID   uint   `json:"domain_id"`
	DomainName string `json:"domain_name"`
	Status     string `json:"verify_status"`
	Message    string `json:"message"`
}

// DomainService 域名业务逻辑层
type DomainService struct {
	repo     *repository.DomainRepository
	certRepo *repository.CertRepository
}

// NewDomainService 创建 DomainService 实例
func NewDomainService(repo *repository.DomainRepository, certRepo *repository.CertRepository) *DomainService {
	return &DomainService{
		repo:     repo,
		certRepo: certRepo,
	}
}

// toVO 将 Domain 转换为 DomainVO
func (s *DomainService) toVO(d *model.Domain) *DomainVO {
	return &DomainVO{
		ID:            d.ID,
		Name:          d.Name,
		CertificateID: d.CertificateID,
		VerifyStatus:  d.VerifyStatus,
		LastCheckAt:   d.LastCheckAt,
		CreatedAt:     d.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     d.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Create 创建域名
func (s *DomainService) Create(name string) (*DomainVO, error) {
	if name == "" {
		return nil, errors.New("domain name is required")
	}

	// 验证域名格式
	if !isValidDomain(name) {
		return nil, errors.New("invalid domain format")
	}

	// 检查域名是否已存在
	if _, err := s.repo.GetByName(name); err == nil {
		return nil, fmt.Errorf("domain %s already exists", name)
	}

	domain := &model.Domain{
		Name:         name,
		VerifyStatus: VerifyStatusUnchecked,
	}

	if err := s.repo.Create(domain); err != nil {
		return nil, fmt.Errorf("failed to create domain: %w", err)
	}

	return s.toVO(domain), nil
}

// Update 更新域名（关联/取消关联证书）
func (s *DomainService) Update(id uint, certificateID *uint) (*DomainVO, error) {
	domain, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	// 如果传入了 certificateID，验证证书是否存在
	if certificateID != nil && *certificateID > 0 {
		if _, err := s.certRepo.GetByID(*certificateID); err != nil {
			return nil, fmt.Errorf("certificate not found: %w", err)
		}
		domain.CertificateID = *certificateID
	} else if certificateID != nil && *certificateID == 0 {
		// 传入 0 表示取消关联
		domain.CertificateID = 0
	}

	if err := s.repo.Update(domain); err != nil {
		return nil, fmt.Errorf("failed to update domain: %w", err)
	}

	return s.toVO(domain), nil
}

// Delete 删除域名
func (s *DomainService) Delete(id uint) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}
	return s.repo.Delete(id)
}

// Get 获取域名详情
func (s *DomainService) Get(id uint) (*DomainVO, error) {
	domain, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}
	return s.toVO(domain), nil
}

// List 分页获取域名列表
func (s *DomainService) List(page, pageSize int, search string) ([]*DomainVO, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	domains, total, err := s.repo.List(page, pageSize, search)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list domains: %w", err)
	}

	vos := make([]*DomainVO, 0, len(domains))
	for i := range domains {
		vos = append(vos, s.toVO(&domains[i]))
	}

	return vos, total, nil
}

// VerifyDomainCert 远端校验域名证书，对比本地关联证书，更新 verify_status/last_check_at
func (s *DomainService) VerifyDomainCert(id uint) (*VerifyResult, error) {
	domain, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	result := &VerifyResult{
		DomainID:   domain.ID,
		DomainName: domain.Name,
	}

	// 1. 获取远端证书信息
	remoteCert, err := certutil.VerifyRemoteCert(domain.Name)
	if err != nil {
		result.Status = VerifyStatusError
		result.Message = fmt.Sprintf("failed to verify remote certificate: %v", err)
		s.updateVerifyStatus(domain.ID, VerifyStatusError)
		return result, nil
	}

	// 2. 检查远端证书是否过期
	if remoteCert.IsExpired {
		result.Status = VerifyStatusExpired
		result.Message = fmt.Sprintf("remote certificate has expired on %s", remoteCert.NotAfter.Format("2006-01-02"))
		s.updateVerifyStatus(domain.ID, VerifyStatusExpired)
		return result, nil
	}

	// 3. 如果有关联的本地证书，进行对比
	if domain.CertificateID > 0 {
		localCert, err := s.certRepo.GetByID(domain.CertificateID)
		if err != nil {
			result.Status = VerifyStatusError
			result.Message = fmt.Sprintf("failed to get local certificate: %v", err)
			s.updateVerifyStatus(domain.ID, VerifyStatusError)
			return result, nil
		}

		// 对比指纹
		if localCert.Fingerprint != "" && remoteCert.Fingerprint != localCert.Fingerprint {
			result.Status = VerifyStatusMismatch
			result.Message = "certificate fingerprint mismatch: local certificate does not match remote"
			s.updateVerifyStatus(domain.ID, VerifyStatusMismatch)
			return result, nil
		}

		// 对比序列号
		if localCert.SerialNumber != "" && remoteCert.SerialNumber != localCert.SerialNumber {
			result.Status = VerifyStatusMismatch
			result.Message = "certificate serial number mismatch: local certificate does not match remote"
			s.updateVerifyStatus(domain.ID, VerifyStatusMismatch)
			return result, nil
		}

		// 对比过期时间（允许 1 分钟的误差）
		if !localCert.ExpireAt.IsZero() {
			timeDiff := remoteCert.NotAfter.Sub(localCert.ExpireAt)
			if timeDiff < 0 {
				timeDiff = -timeDiff
			}
			if timeDiff > time.Minute {
				result.Status = VerifyStatusMismatch
				result.Message = fmt.Sprintf("certificate expiry mismatch: local=%s, remote=%s",
					localCert.ExpireAt.Format("2006-01-02"), remoteCert.NotAfter.Format("2006-01-02"))
				s.updateVerifyStatus(domain.ID, VerifyStatusMismatch)
				return result, nil
			}
		}
	}

	// 4. 校验通过
	result.Status = VerifyStatusNormal
	result.Message = fmt.Sprintf("certificate is valid, expires in %d days", remoteCert.DaysUntilExpiry)
	s.updateVerifyStatus(domain.ID, VerifyStatusNormal)

	return result, nil
}

// BatchVerify 批量校验域名证书
func (s *DomainService) BatchVerify(ids []uint) ([]*VerifyResult, error) {
	if len(ids) == 0 {
		return []*VerifyResult{}, nil
	}

	results := make([]*VerifyResult, 0, len(ids))
	for _, id := range ids {
		result, err := s.VerifyDomainCert(id)
		if err != nil {
			// 单个域名校验失败不影响其他域名
			result = &VerifyResult{
				DomainID: id,
				Status:   VerifyStatusError,
				Message:  err.Error(),
			}
		}
		results = append(results, result)
	}

	return results, nil
}

// updateVerifyStatus 更新域名校验状态
func (s *DomainService) updateVerifyStatus(id uint, status string) {
	now := time.Now()
	domain := &model.Domain{
		VerifyStatus: status,
		LastCheckAt:  &now,
	}
	domain.ID = id
	s.repo.Update(domain)
}

// isValidDomain 验证域名格式
func isValidDomain(domain string) bool {
	if domain == "" {
		return false
	}

	// 去除可能的端口
	if strings.Contains(domain, ":") {
		host, _, _ := strings.Cut(domain, ":")
		domain = host
	}

	// 基本验证：域名不能为空，不能包含空格
	if strings.Contains(domain, " ") {
		return false
	}

	// 至少包含一个点（顶级域名除外，但这里要求至少二级域名）
	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}
