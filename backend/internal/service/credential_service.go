package service

import (
	"errors"
	"fmt"
	"strings"

	"certmanager-backend/internal/model"
	"certmanager-backend/internal/repository"
	"certmanager-backend/pkg/crypto"
)

// validProviderTypes 支持的云提供商类型
var validProviderTypes = map[string]bool{
	"aliyun":     true,
	"tencent":    true,
	"volcengine": true,
	"wangsu":     true,
	"aws":        true,
	"azure":      true,
}

// CredentialVO 凭证视图对象（脱敏后）
type CredentialVO struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	ProviderType string `json:"provider_type"`
	AccessKey    string `json:"access_key"`
	ExtraConfig  string `json:"extra_config"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// CredentialService 云凭证业务逻辑层
type CredentialService struct {
	repo   *repository.CredentialRepository
	aesKey string
}

// NewCredentialService 创建 CredentialService 实例
func NewCredentialService(repo *repository.CredentialRepository, aesKey string) *CredentialService {
	return &CredentialService{repo: repo, aesKey: aesKey}
}

// maskAccessKey 脱敏 accessKey：只显示前4位+****
func maskAccessKey(accessKey string) string {
	if len(accessKey) <= 4 {
		return "****"
	}
	return accessKey[:4] + "****"
}

// toVO 将 CloudCredential 转换为脱敏的 CredentialVO
func (s *CredentialService) toVO(c *model.CloudCredential) *CredentialVO {
	accessKey := ""
	if c.AccessKeyEncrypted != "" {
		if decrypted, err := crypto.Decrypt(c.AccessKeyEncrypted, s.aesKey); err == nil {
			accessKey = maskAccessKey(decrypted)
		}
	}
	return &CredentialVO{
		ID:           c.ID,
		Name:         c.Name,
		ProviderType: c.ProviderType,
		AccessKey:    accessKey,
		ExtraConfig:  c.ExtraConfig,
		Status:       c.Status,
		CreatedAt:    c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Create 创建凭证
func (s *CredentialService) Create(name, providerType, accessKey, secretKey, extraConfig string) (*CredentialVO, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if !validProviderTypes[providerType] {
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}

	encryptedAccessKey, err := crypto.Encrypt(accessKey, s.aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt access key: %w", err)
	}

	encryptedSecretKey, err := crypto.Encrypt(secretKey, s.aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret key: %w", err)
	}

	credential := &model.CloudCredential{
		Name:               name,
		ProviderType:       providerType,
		AccessKeyEncrypted: encryptedAccessKey,
		SecretKeyEncrypted: encryptedSecretKey,
		ExtraConfig:        extraConfig,
		Status:             "active",
	}

	if err := s.repo.Create(credential); err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	return s.toVO(credential), nil
}

// Update 更新凭证，accessKey/secretKey 为空时不更新
func (s *CredentialService) Update(id uint, name, providerType, accessKey, secretKey, extraConfig string) (*CredentialVO, error) {
	credential, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	if name != "" {
		credential.Name = name
	}
	if providerType != "" {
		if !validProviderTypes[providerType] {
			return nil, fmt.Errorf("unsupported provider type: %s", providerType)
		}
		credential.ProviderType = providerType
	}
	if accessKey != "" {
		encryptedAccessKey, err := crypto.Encrypt(accessKey, s.aesKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt access key: %w", err)
		}
		credential.AccessKeyEncrypted = encryptedAccessKey
	}
	if secretKey != "" {
		encryptedSecretKey, err := crypto.Encrypt(secretKey, s.aesKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt secret key: %w", err)
		}
		credential.SecretKeyEncrypted = encryptedSecretKey
	}
	if extraConfig != "" {
		credential.ExtraConfig = extraConfig
	}

	if err := s.repo.Update(credential); err != nil {
		return nil, fmt.Errorf("failed to update credential: %w", err)
	}

	return s.toVO(credential), nil
}

// Delete 删除凭证
func (s *CredentialService) Delete(id uint) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return fmt.Errorf("credential not found: %w", err)
	}
	return s.repo.Delete(id)
}

// Get 获取凭证详情（脱敏）
func (s *CredentialService) Get(id uint) (*CredentialVO, error) {
	credential, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}
	return s.toVO(credential), nil
}

// List 分页获取凭证列表（脱敏）
func (s *CredentialService) List(page, pageSize int, providerType string) ([]*CredentialVO, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	credentials, total, err := s.repo.List(page, pageSize, providerType)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list credentials: %w", err)
	}

	vos := make([]*CredentialVO, 0, len(credentials))
	for i := range credentials {
		vos = append(vos, s.toVO(&credentials[i]))
	}

	return vos, total, nil
}

// TestConnection 测试连接：模拟测试，检查凭证格式有效性
func (s *CredentialService) TestConnection(id uint) (string, string) {
	credential, err := s.repo.GetByID(id)
	if err != nil {
		return "failed", "credential not found"
	}

	// 解密 accessKey/secretKey 验证格式
	accessKey, err := crypto.Decrypt(credential.AccessKeyEncrypted, s.aesKey)
	if err != nil || strings.TrimSpace(accessKey) == "" {
		return "failed", "invalid access key"
	}

	secretKey, err := crypto.Decrypt(credential.SecretKeyEncrypted, s.aesKey)
	if err != nil || strings.TrimSpace(secretKey) == "" {
		return "failed", "invalid secret key"
	}

	// 模拟各云商格式校验
	switch credential.ProviderType {
	case "aliyun":
		if len(accessKey) < 16 {
			return "failed", "aliyun access key format invalid"
		}
	case "tencent":
		if len(accessKey) < 16 {
			return "failed", "tencent access key format invalid"
		}
	case "aws":
		if len(accessKey) < 16 {
			return "failed", "aws access key format invalid"
		}
	}

	return "success", "connection test passed"
}
