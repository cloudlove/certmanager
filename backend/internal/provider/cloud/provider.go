package cloud

import (
	"context"
	"fmt"
	"strings"
)

// DeployRequest 部署请求结构体
type DeployRequest struct {
	CertPEM       string
	ChainPEM      string
	PrivateKeyPEM string
	ResourceID    string
	ResourceType  string // CDN, SLB, CLB, DCDN, CloudFront, ELB, AppGateway
}

// DeployResponse 部署响应结构体
type DeployResponse struct {
	TaskID  string
	Status  string
	Message string
}

// RollbackRequest 回滚请求结构体
type RollbackRequest struct {
	ResourceID   string
	ResourceType string
	OldCertInfo  string
}

// CloudDeployProvider 云部署 Provider 接口
type CloudDeployProvider interface {
	// DeployCert 部署证书到云资源
	DeployCert(ctx context.Context, req DeployRequest) (*DeployResponse, error)
	// GetDeployStatus 获取部署任务状态
	GetDeployStatus(ctx context.Context, taskID string) (string, error)
	// GetCurrentCert 获取当前资源上的证书信息（用于快照）
	GetCurrentCert(ctx context.Context, resourceID, resourceType string) (string, error)
	// Rollback 回滚证书
	Rollback(ctx context.Context, req RollbackRequest) error
}

// NewCloudDeployProvider 创建云部署 Provider 工厂函数
func NewCloudDeployProvider(providerType string, accessKey, secretKey string) (CloudDeployProvider, error) {
	switch providerType {
	case "aliyun":
		return NewAliyunProvider(accessKey, secretKey), nil
	case "tencent":
		return NewTencentProvider(accessKey, secretKey), nil
	case "volcengine":
		return NewVolcengineProvider(accessKey, secretKey), nil
	case "wangsu":
		return NewWangsuProvider(accessKey, secretKey), nil
	case "aws":
		return NewAWSProvider(accessKey, secretKey), nil
	case "azure":
		return NewAzureProvider(accessKey, secretKey), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

// IsValidResourceType 检查资源类型是否有效
func IsValidResourceType(providerType, resourceType string) bool {
	resourceType = strings.ToLower(resourceType)
	validTypes := map[string][]string{
		"aliyun":     {"cdn", "slb", "dcdn"},
		"tencent":    {"cdn", "clb"},
		"volcengine": {"cdn", "clb"},
		"wangsu":     {"cdn"},
		"aws":        {"cloudfront", "elb"},
		"azure":      {"cdn", "appgateway"},
	}

	types, ok := validTypes[providerType]
	if !ok {
		return false
	}

	for _, t := range types {
		if t == resourceType {
			return true
		}
	}
	return false
}
