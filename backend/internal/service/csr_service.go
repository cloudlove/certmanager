package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"certmanager-backend/internal/model"
	"certmanager-backend/internal/repository"
	"certmanager-backend/pkg/crypto"
)

// CSRVO CSR 视图对象（不含敏感信息）
type CSRVO struct {
	ID           uint     `json:"id"`
	CommonName   string   `json:"common_name"`
	SAN          []string `json:"san"`
	KeyAlgorithm string   `json:"key_algorithm"`
	KeySize      string   `json:"key_size"`
	Status       string   `json:"status"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
	// 阿里云 CreateCsr 参数
	CountryCode string `json:"country_code"`
	Province    string `json:"province"`
	Locality    string `json:"locality"`
	CorpName    string `json:"corp_name"`
	Department  string `json:"department"`
}

// CSRDetailVO CSR 详情视图对象
type CSRDetailVO struct {
	ID           uint     `json:"id"`
	CommonName   string   `json:"common_name"`
	SAN          []string `json:"san"`
	KeyAlgorithm string   `json:"key_algorithm"`
	KeySize      string   `json:"key_size"`
	CSRPEM       string   `json:"csr_pem"`
	Status       string   `json:"status"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
	// 阿里云 CreateCsr 参数
	CountryCode string `json:"country_code"`
	Province    string `json:"province"`
	Locality    string `json:"locality"`
	CorpName    string `json:"corp_name"`
	Department  string `json:"department"`
}

// CSRService CSR 业务逻辑层
type CSRService struct {
	repo   *repository.CSRRepository
	aesKey string
}

// NewCSRService 创建 CSRService 实例
func NewCSRService(repo *repository.CSRRepository, aesKey string) *CSRService {
	return &CSRService{repo: repo, aesKey: aesKey}
}

// toVO 将 CSRRecord 转换为 CSRVO
func (s *CSRService) toVO(c *model.CSRRecord) *CSRVO {
	return &CSRVO{
		ID:           c.ID,
		CommonName:   c.CommonName,
		SAN:          splitSAN(c.SAN),
		KeyAlgorithm: c.KeyAlgorithm,
		KeySize:      strconv.Itoa(c.KeySize),
		Status:       c.Status,
		CreatedAt:    c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    c.UpdatedAt.Format("2006-01-02 15:04:05"),
		CountryCode:  c.CountryCode,
		Province:     c.Province,
		Locality:     c.Locality,
		CorpName:     c.CorpName,
		Department:   c.Department,
	}
}

// toDetailVO 将 CSRRecord 转换为 CSRDetailVO
func (s *CSRService) toDetailVO(c *model.CSRRecord) *CSRDetailVO {
	return &CSRDetailVO{
		ID:           c.ID,
		CommonName:   c.CommonName,
		SAN:          splitSAN(c.SAN),
		KeyAlgorithm: c.KeyAlgorithm,
		KeySize:      strconv.Itoa(c.KeySize),
		CSRPEM:       c.CSRPEM,
		Status:       c.Status,
		CreatedAt:    c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    c.UpdatedAt.Format("2006-01-02 15:04:05"),
		CountryCode:  c.CountryCode,
		Province:     c.Province,
		Locality:     c.Locality,
		CorpName:     c.CorpName,
		Department:   c.Department,
	}
}

// splitSAN 将逗号分隔的 SAN 字符串分割为切片
func splitSAN(san string) []string {
	if san == "" {
		return []string{}
	}
	parts := strings.Split(san, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// joinSAN 将 SAN 切片连接为逗号分隔的字符串
func joinSAN(sans []string) string {
	var validSANs []string
	for _, san := range sans {
		san = strings.TrimSpace(san)
		if san != "" {
			validSANs = append(validSANs, san)
		}
	}
	return strings.Join(validSANs, ",")
}

// getKeySizeInt 将密钥大小字符串转换为整数
func getKeySizeInt(keyAlgorithm, keySize string) int {
	switch strings.ToUpper(keyAlgorithm) {
	case "RSA":
		switch keySize {
		case "2048":
			return 2048
		case "4096":
			return 4096
		default:
			return 2048
		}
	case "ECC":
		switch keySize {
		case "P256", "256":
			return 256
		case "P384", "384":
			return 384
		default:
			return 256
		}
	}
	return 2048
}

// Generate 生成 CSR，加密存储私钥
func (s *CSRService) Generate(commonName string, sans []string, keyAlgorithm string, keySize int, countryCode, province, locality, corpName, department string) (*CSRVO, error) {
	if commonName == "" {
		return nil, errors.New("common name is required")
	}

	// 验证密钥算法
	keyAlgorithm = strings.ToUpper(keyAlgorithm)
	if keyAlgorithm != "RSA" && keyAlgorithm != "ECC" {
		return nil, errors.New("key algorithm must be RSA or ECC")
	}

	// 验证密钥大小
	if keySize == 0 {
		if keyAlgorithm == "RSA" {
			keySize = 2048
		} else {
			keySize = 384
		}
	}

	// 生成 CSR 和私钥
	csrPEM, privateKeyPEM, err := crypto.GenerateCSR(commonName, sans, keyAlgorithm, strconv.Itoa(keySize), countryCode, province, locality, corpName, department)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CSR: %w", err)
	}

	// 加密私钥
	encryptedPrivateKey, err := crypto.Encrypt(privateKeyPEM, s.aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// 创建记录
	csr := &model.CSRRecord{
		CommonName:          commonName,
		SAN:                 joinSAN(sans),
		KeyAlgorithm:        keyAlgorithm,
		KeySize:             getKeySizeInt(keyAlgorithm, strconv.Itoa(keySize)),
		CSRPEM:              csrPEM,
		PrivateKeyEncrypted: encryptedPrivateKey,
		Status:              "active",
		CountryCode:         countryCode,
		Province:            province,
		Locality:            locality,
		CorpName:            corpName,
		Department:          department,
	}

	if err := s.repo.Create(csr); err != nil {
		return nil, fmt.Errorf("failed to save CSR: %w", err)
	}

	return s.toVO(csr), nil
}

// Get 获取 CSR 详情（不含私钥）
func (s *CSRService) Get(id uint) (*CSRDetailVO, error) {
	csr, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("CSR not found: %w", err)
	}
	return s.toDetailVO(csr), nil
}

// List 分页获取 CSR 列表
func (s *CSRService) List(page, pageSize int, search string) ([]*CSRVO, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	csrs, total, err := s.repo.List(page, pageSize, search)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list CSRs: %w", err)
	}

	vos := make([]*CSRVO, 0, len(csrs))
	for i := range csrs {
		vos = append(vos, s.toVO(&csrs[i]))
	}

	return vos, total, nil
}

// Delete 删除 CSR
func (s *CSRService) Delete(id uint) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return fmt.Errorf("CSR not found: %w", err)
	}
	return s.repo.Delete(id)
}

// Parse 解析外部提供的 CSR
func (s *CSRService) Parse(csrPEM string) (*crypto.CSRInfo, error) {
	if strings.TrimSpace(csrPEM) == "" {
		return nil, errors.New("CSR PEM is required")
	}
	return crypto.ParseCSR(csrPEM)
}

// DownloadCSR 返回 CSR PEM 内容
func (s *CSRService) DownloadCSR(id uint) (string, error) {
	csr, err := s.repo.GetByID(id)
	if err != nil {
		return "", fmt.Errorf("CSR not found: %w", err)
	}
	return csr.CSRPEM, nil
}

// DownloadPrivateKey 解密返回私钥 PEM
func (s *CSRService) DownloadPrivateKey(id uint) (string, error) {
	csr, err := s.repo.GetByID(id)
	if err != nil {
		return "", fmt.Errorf("CSR not found: %w", err)
	}

	if csr.PrivateKeyEncrypted == "" {
		return "", errors.New("private key not found")
	}

	privateKeyPEM, err := crypto.Decrypt(csr.PrivateKeyEncrypted, s.aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt private key: %w", err)
	}

	return privateKeyPEM, nil
}
