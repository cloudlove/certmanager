package ca

import (
	"context"
	"fmt"
)

// ApplyCertRequest 申请证书请求
type ApplyCertRequest struct {
	Domain        string
	CSRPEM        string
	ValidityYears int
	ProductType   string // DV/OV/EV
	ValidateType  string // DNS 或 FILE
}

// ApplyCertResponse 申请证书响应
type ApplyCertResponse struct {
	OrderID string
	Status  string
}

// CertStatusResponse 证书状态响应
type CertStatusResponse struct {
	Status  string // domain_verify / process / issued / failed
	OrderID string
	// DNS验证信息
	RecordDomain string // DNS记录主机名，如 _dnsauth
	RecordType   string // DNS记录类型，如 TXT
	RecordValue  string // DNS记录值
	// 文件验证信息
	Uri     string // 验证文件路径
	Content string // 验证文件内容
	// 已签发时的证书内容
	Certificate string // 证书PEM
	PrivateKey  string // 私钥PEM
}

// CAProvider CA 提供商接口
type CAProvider interface {
	ApplyCert(ctx context.Context, req ApplyCertRequest) (*ApplyCertResponse, error)
	GetCertStatus(ctx context.Context, orderID string) (*CertStatusResponse, error)
	DownloadCert(ctx context.Context, orderID string) (certPEM string, chainPEM string, err error)
}

// NewCAProvider 创建 CA Provider 实例
func NewCAProvider(providerType string, accessKey, secretKey string) (CAProvider, error) {
	switch providerType {
	case "aliyun":
		return NewAliyunProvider(accessKey, secretKey), nil
	case "tencent":
		return NewTencentProvider(accessKey, secretKey), nil
	case "volcengine":
		return NewVolcengineProvider(accessKey, secretKey), nil
	default:
		return nil, fmt.Errorf("unsupported CA provider type: %s", providerType)
	}
}
