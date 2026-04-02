package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"certmanager-backend/internal/model"
	"certmanager-backend/internal/provider/ca"
	"certmanager-backend/internal/repository"
	"certmanager-backend/pkg/certutil"
	"certmanager-backend/pkg/crypto"
)

// CertificateVO 证书视图对象（不含敏感信息）
type CertificateVO struct {
	ID           uint   `json:"id"`
	Domain       string `json:"domain"`
	CAProvider   string `json:"ca_provider"`
	Status       string `json:"status"`
	ExpireAt     string `json:"expire_at"`
	Issuer       string `json:"issuer"`
	Fingerprint  string `json:"fingerprint"`
	KeyAlgorithm string `json:"key_algorithm"`
	SerialNumber string `json:"serial_number"`
	OrderID      string `json:"order_id"`
	VerifyType   string `json:"verify_type"`
	VerifyInfo   string `json:"verify_info"`
	ProductType  string `json:"product_type"`
	DomainType   string `json:"domain_type"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// CertificateDetailVO 证书详情视图对象
type CertificateDetailVO struct {
	ID           uint   `json:"id"`
	Domain       string `json:"domain"`
	CAProvider   string `json:"ca_provider"`
	Status       string `json:"status"`
	ExpireAt     string `json:"expire_at"`
	CSRID        uint   `json:"csr_id"`
	CertPEM      string `json:"cert_pem"`
	ChainPEM     string `json:"chain_pem"`
	Issuer       string `json:"issuer"`
	Fingerprint  string `json:"fingerprint"`
	KeyAlgorithm string `json:"key_algorithm"`
	SerialNumber string `json:"serial_number"`
	OrderID      string `json:"order_id"`
	VerifyType   string `json:"verify_type"`
	VerifyInfo   string `json:"verify_info"`
	ProductType  string `json:"product_type"`
	DomainType   string `json:"domain_type"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// CertificateService 证书业务逻辑层
type CertificateService struct {
	certRepo       *repository.CertificateRepository
	csrRepo        *repository.CSRRepository
	credentialRepo *repository.CredentialRepository
	aesKey         string
}

// NewCertificateService 创建 CertificateService 实例
func NewCertificateService(certRepo *repository.CertificateRepository, csrRepo *repository.CSRRepository, credentialRepo *repository.CredentialRepository, aesKey string) *CertificateService {
	return &CertificateService{
		certRepo:       certRepo,
		csrRepo:        csrRepo,
		credentialRepo: credentialRepo,
		aesKey:         aesKey,
	}
}

// toVO 将 Certificate 转换为 CertificateVO
func (s *CertificateService) toVO(c *model.Certificate) *CertificateVO {
	return &CertificateVO{
		ID:           c.ID,
		Domain:       c.Domain,
		CAProvider:   c.CAProvider,
		Status:       c.Status,
		ExpireAt:     c.ExpireAt.Format("2006-01-02 15:04:05"),
		Issuer:       c.Issuer,
		Fingerprint:  c.Fingerprint,
		KeyAlgorithm: c.KeyAlgorithm,
		SerialNumber: c.SerialNumber,
		OrderID:      c.OrderID,
		VerifyType:   c.VerifyType,
		VerifyInfo:   c.VerifyInfo,
		ProductType:  c.ProductType,
		DomainType:   c.DomainType,
		CreatedAt:    c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// toDetailVO 将 Certificate 转换为 CertificateDetailVO
func (s *CertificateService) toDetailVO(c *model.Certificate) *CertificateDetailVO {
	return &CertificateDetailVO{
		ID:           c.ID,
		Domain:       c.Domain,
		CAProvider:   c.CAProvider,
		Status:       c.Status,
		ExpireAt:     c.ExpireAt.Format("2006-01-02 15:04:05"),
		CSRID:        c.CSRID,
		CertPEM:      c.CertPEM,
		ChainPEM:     c.ChainPEM,
		Issuer:       c.Issuer,
		Fingerprint:  c.Fingerprint,
		KeyAlgorithm: c.KeyAlgorithm,
		SerialNumber: c.SerialNumber,
		OrderID:      c.OrderID,
		VerifyType:   c.VerifyType,
		VerifyInfo:   c.VerifyInfo,
		ProductType:  c.ProductType,
		DomainType:   c.DomainType,
		CreatedAt:    c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ApplyCertificate 申请证书（异步流程）
// 提交申请后不等待签发，而是返回验证信息供用户配置
func (s *CertificateService) ApplyCertificate(caProvider string, domain string, csrID uint, credentialID uint, validateType string, productType string, domainType string) (*CertificateVO, error) {
	// 验证参数
	if caProvider == "" || domain == "" {
		return nil, errors.New("ca provider and domain are required")
	}
	if csrID == 0 {
		return nil, errors.New("csr id is required")
	}
	if credentialID == 0 {
		return nil, errors.New("credential id is required")
	}
	if validateType == "" {
		validateType = "DNS" // 默认 DNS 验证
	}

	// 获取 CSR 记录
	csr, err := s.csrRepo.GetByID(csrID)
	if err != nil {
		return nil, fmt.Errorf("CSR not found: %w", err)
	}

	// 获取凭证
	credential, err := s.credentialRepo.GetByID(credentialID)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	// 解密凭证
	accessKey, err := crypto.Decrypt(credential.AccessKeyEncrypted, s.aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt access key: %w", err)
	}
	secretKey, err := crypto.Decrypt(credential.SecretKeyEncrypted, s.aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret key: %w", err)
	}

	// 创建 CA Provider
	provider, err := ca.NewCAProvider(caProvider, accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA provider: %w", err)
	}

	// 调用 provider.ApplyCert() 提交申请
	ctx := context.Background()
	applyReq := ca.ApplyCertRequest{
		Domain:       domain,
		CSRPEM:       csr.CSRPEM,
		ProductType:  productType,
		ValidateType: validateType,
	}

	applyResp, err := provider.ApplyCert(ctx, applyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to apply certificate: %w", err)
	}

	// 调用 provider.GetCertStatus() 获取验证信息
	statusResp, err := provider.GetCertStatus(ctx, applyResp.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cert status: %w", err)
	}

	// 将验证信息序列化为 JSON 存入 VerifyInfo
	var verifyInfo string
	if validateType == "DNS" {
		verifyInfo = fmt.Sprintf(`{"recordDomain":"%s","recordType":"%s","recordValue":"%s"}`,
			statusResp.RecordDomain, statusResp.RecordType, statusResp.RecordValue)
	} else {
		verifyInfo = fmt.Sprintf(`{"uri":"%s","content":"%s"}`,
			statusResp.Uri, statusResp.Content)
	}

	// 创建 Certificate 记录保存到数据库
	cert := &model.Certificate{
		Domain:       domain,
		CAProvider:   caProvider,
		Status:       statusResp.Status,
		CSRID:        csrID,
		CredentialID: credentialID,
		OrderID:      applyResp.OrderID,
		VerifyType:   validateType,
		VerifyInfo:   verifyInfo,
		ProductType:  productType,
		DomainType:   domainType,
	}

	// 使用 CSR 的私钥
	cert.PrivateKeyEncrypted = csr.PrivateKeyEncrypted

	if err := s.certRepo.Create(cert); err != nil {
		return nil, fmt.Errorf("failed to save certificate: %w", err)
	}

	return s.toVO(cert), nil
}

// ImportCertificate 导入已有证书
func (s *CertificateService) ImportCertificate(certPEM, chainPEM, privateKeyPEM string) (*CertificateVO, error) {
	if certPEM == "" {
		return nil, errors.New("certificate PEM is required")
	}

	// 解析证书信息
	certInfo, err := certutil.ParseCertificate(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// 加密私钥
	var encryptedPrivateKey string
	if privateKeyPEM != "" {
		encryptedPrivateKey, err = crypto.Encrypt(privateKeyPEM, s.aesKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt private key: %w", err)
		}
	}

	// 确定证书状态
	status := "issued"
	if certInfo.IsExpired {
		status = "expired"
	}

	// 创建证书记录
	cert := &model.Certificate{
		Domain:              certInfo.Domain,
		CAProvider:          "imported",
		Status:              status,
		ExpireAt:            certInfo.NotAfter,
		CertPEM:             certPEM,
		ChainPEM:            chainPEM,
		PrivateKeyEncrypted: encryptedPrivateKey,
		Issuer:              certInfo.Issuer,
		Fingerprint:         certInfo.Fingerprint,
		KeyAlgorithm:        certInfo.KeyAlgorithm,
		SerialNumber:        certInfo.SerialNumber,
	}

	if err := s.certRepo.Create(cert); err != nil {
		return nil, fmt.Errorf("failed to save certificate: %w", err)
	}

	return s.toVO(cert), nil
}

// Get 获取证书详情
func (s *CertificateService) Get(id uint) (*CertificateDetailVO, error) {
	cert, err := s.certRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("certificate not found: %w", err)
	}
	return s.toDetailVO(cert), nil
}

// List 分页获取证书列表
func (s *CertificateService) List(page, pageSize int, status, search, sortBy string) ([]*CertificateVO, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	certs, total, err := s.certRepo.List(page, pageSize, status, search, sortBy)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list certificates: %w", err)
	}

	vos := make([]*CertificateVO, 0, len(certs))
	for i := range certs {
		vos = append(vos, s.toVO(&certs[i]))
	}

	return vos, total, nil
}

// Delete 删除证书
func (s *CertificateService) Delete(id uint) error {
	if _, err := s.certRepo.GetByID(id); err != nil {
		return fmt.Errorf("certificate not found: %w", err)
	}
	return s.certRepo.Delete(id)
}

// SyncCertStatus 同步单个证书状态（手动触发）
func (s *CertificateService) SyncCertStatus(id uint) (*CertificateDetailVO, error) {
	cert, err := s.certRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("certificate not found: %w", err)
	}

	// 如果是导入的证书，不支持同步
	if cert.CAProvider == "imported" {
		return nil, errors.New("imported certificate does not support sync")
	}

	// 如果没有 OrderID，无法同步
	if cert.OrderID == "" {
		return nil, errors.New("certificate has no order id, cannot sync")
	}

	// 如果已经签发，只检查过期
	if cert.Status == "issued" {
		now := time.Now()
		if now.After(cert.ExpireAt) {
			cert.Status = "expired"
			if err := s.certRepo.Update(cert); err != nil {
				return nil, fmt.Errorf("failed to update certificate status: %w", err)
			}
		}
		return s.toDetailVO(cert), nil
	}

	// 获取凭证
	if cert.CredentialID == 0 {
		return nil, errors.New("certificate has no credential id, cannot sync")
	}
	credential, err := s.credentialRepo.GetByID(cert.CredentialID)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	// 解密凭证
	accessKey, err := crypto.Decrypt(credential.AccessKeyEncrypted, s.aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt access key: %w", err)
	}
	secretKey, err := crypto.Decrypt(credential.SecretKeyEncrypted, s.aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret key: %w", err)
	}

	// 创建 CA Provider
	provider, err := ca.NewCAProvider(cert.CAProvider, accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA provider: %w", err)
	}

	// 调用 provider.GetCertStatus() 获取状态
	ctx := context.Background()
	statusResp, err := provider.GetCertStatus(ctx, cert.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cert status: %w", err)
	}

	// 根据状态更新证书
	s.updateCertFromStatus(cert, statusResp)

	if err := s.certRepo.Update(cert); err != nil {
		return nil, fmt.Errorf("failed to update certificate: %w", err)
	}

	return s.toDetailVO(cert), nil
}

// SyncPendingCertificates 同步所有待处理证书状态（定时任务调用）
func (s *CertificateService) SyncPendingCertificates() {
	// 查询所有待处理状态的证书
	statuses := []string{"domain_verify", "process"}
	certs, err := s.certRepo.ListByStatuses(statuses)
	if err != nil {
		log.Printf("Failed to list pending certificates: %v", err)
		return
	}

	if len(certs) == 0 {
		return
	}

	log.Printf("Syncing %d pending certificates...", len(certs))

	for i := range certs {
		cert := &certs[i]

		// 如果没有 OrderID，跳过
		if cert.OrderID == "" {
			continue
		}

		// 如果没有 CredentialID，跳过
		if cert.CredentialID == 0 {
			log.Printf("Certificate %d has no credential id, skip", cert.ID)
			continue
		}

		// 获取凭证
		credential, err := s.credentialRepo.GetByID(cert.CredentialID)
		if err != nil {
			log.Printf("Failed to get credential for certificate %d: %v", cert.ID, err)
			continue
		}

		// 解密凭证
		accessKey, err := crypto.Decrypt(credential.AccessKeyEncrypted, s.aesKey)
		if err != nil {
			log.Printf("Failed to decrypt access key for certificate %d: %v", cert.ID, err)
			continue
		}
		secretKey, err := crypto.Decrypt(credential.SecretKeyEncrypted, s.aesKey)
		if err != nil {
			log.Printf("Failed to decrypt secret key for certificate %d: %v", cert.ID, err)
			continue
		}

		// 创建 CA Provider
		provider, err := ca.NewCAProvider(cert.CAProvider, accessKey, secretKey)
		if err != nil {
			log.Printf("Failed to create CA provider for certificate %d: %v", cert.ID, err)
			continue
		}

		// 调用 provider.GetCertStatus() 获取状态
		ctx := context.Background()
		statusResp, err := provider.GetCertStatus(ctx, cert.OrderID)
		if err != nil {
			log.Printf("Failed to get cert status for certificate %d: %v", cert.ID, err)
			continue
		}

		// 根据状态更新证书
		s.updateCertFromStatus(cert, statusResp)

		if err := s.certRepo.Update(cert); err != nil {
			log.Printf("Failed to update certificate %d: %v", cert.ID, err)
			continue
		}

		log.Printf("Certificate %d status updated to: %s", cert.ID, cert.Status)
	}

	log.Println("Pending certificates sync completed")
}

// updateCertFromStatus 根据状态响应更新证书
func (s *CertificateService) updateCertFromStatus(cert *model.Certificate, statusResp *ca.CertStatusResponse) {
	switch statusResp.Status {
	case "issued", "certificate":
		// 证书已签发
		cert.Status = "issued"

		// 如果有证书内容，解析并保存
		if statusResp.Certificate != "" {
			cert.CertPEM = statusResp.Certificate

			// 解析证书信息
			certInfo, err := certutil.ParseCertificate(statusResp.Certificate)
			if err == nil {
				cert.ExpireAt = certInfo.NotAfter
				cert.Issuer = certInfo.Issuer
				cert.Fingerprint = certInfo.Fingerprint
				cert.KeyAlgorithm = certInfo.KeyAlgorithm
				cert.SerialNumber = certInfo.SerialNumber
			}
		}

		// 如果有私钥，加密保存
		if statusResp.PrivateKey != "" {
			encryptedKey, err := crypto.Encrypt(statusResp.PrivateKey, s.aesKey)
			if err == nil {
				cert.PrivateKeyEncrypted = encryptedKey
			}
		}

	case "process":
		cert.Status = "process"

	case "failed", "verify_fail":
		cert.Status = "failed"

	case "domain_verify":
		// 更新验证信息（可能变化）
		cert.Status = "domain_verify"
		if cert.VerifyType == "DNS" {
			verifyInfo := map[string]string{
				"recordDomain": statusResp.RecordDomain,
				"recordType":   statusResp.RecordType,
				"recordValue":  statusResp.RecordValue,
			}
			if data, err := json.Marshal(verifyInfo); err == nil {
				cert.VerifyInfo = string(data)
			}
		} else {
			verifyInfo := map[string]string{
				"uri":     statusResp.Uri,
				"content": statusResp.Content,
			}
			if data, err := json.Marshal(verifyInfo); err == nil {
				cert.VerifyInfo = string(data)
			}
		}
	}
}

// CountByStatus 统计各状态证书数量
func (s *CertificateService) CountByStatus() (map[string]int64, error) {
	return s.certRepo.CountByStatus()
}

// DownloadCertificate 下载证书
func (s *CertificateService) DownloadCertificate(id uint) (certPEM, chainPEM string, err error) {
	cert, err := s.certRepo.GetByID(id)
	if err != nil {
		return "", "", fmt.Errorf("certificate not found: %w", err)
	}
	return cert.CertPEM, cert.ChainPEM, nil
}

// DownloadPrivateKey 下载私钥
func (s *CertificateService) DownloadPrivateKey(id uint) (string, error) {
	cert, err := s.certRepo.GetByID(id)
	if err != nil {
		return "", fmt.Errorf("certificate not found: %w", err)
	}

	if cert.PrivateKeyEncrypted == "" {
		return "", errors.New("private key not found")
	}

	privateKeyPEM, err := crypto.Decrypt(cert.PrivateKeyEncrypted, s.aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt private key: %w", err)
	}

	return privateKeyPEM, nil
}
