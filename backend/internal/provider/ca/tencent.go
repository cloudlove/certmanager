package ca

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// TencentProvider 腾讯云 CA Provider
type TencentProvider struct {
	accessKey string
	secretKey string
}

// NewTencentProvider 创建腾讯云 Provider 实例
func NewTencentProvider(accessKey, secretKey string) *TencentProvider {
	return &TencentProvider{
		accessKey: accessKey,
		secretKey: secretKey,
	}
}

// ApplyCert 申请证书（模拟实现）
func (p *TencentProvider) ApplyCert(ctx context.Context, req ApplyCertRequest) (*ApplyCertResponse, error) {
	// 模拟调用腾讯云 SSL 证书 API
	// TODO: 集成腾讯云 SSL SDK
	// 实际实现应调用: https://cloud.tencent.com/document/product/400/41675

	// 生成 mock orderID
	orderID := fmt.Sprintf("tencent-%d-%d", time.Now().Unix(), randInt())

	return &ApplyCertResponse{
		OrderID: orderID,
		Status:  "pending",
	}, nil
}

// GetCertStatus 获取证书状态（模拟实现）
func (p *TencentProvider) GetCertStatus(ctx context.Context, orderID string) (*CertStatusResponse, error) {
	// 模拟调用腾讯云 SSL 证书 API 查询状态
	// TODO: 集成腾讯云 SSL SDK

	return &CertStatusResponse{
		OrderID: orderID,
		Status:  "issued",
	}, nil
}

// DownloadCert 下载证书（模拟实现，返回自签名测试证书）
func (p *TencentProvider) DownloadCert(ctx context.Context, orderID string) (certPEM string, chainPEM string, err error) {
	// 模拟调用腾讯云 SSL 证书 API 下载证书
	// TODO: 集成腾讯云 SSL SDK

	// 生成自签名测试证书
	certPEM, chainPEM, err = generateSelfSignedCert(orderID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate mock certificate: %w", err)
	}

	return certPEM, chainPEM, nil
}

// randInt 生成随机整数
func randInt() int64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return n.Int64()
}

// generateSelfSignedCert 生成自签名测试证书
func generateSelfSignedCert(domain string) (certPEM string, chainPEM string, err error) {
	// 生成私钥
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// 生成序列号
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate serial number: %w", err)
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: domain,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	// 创建自签名证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate: %w", err)
	}

	// 编码证书
	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}))

	// 生成 CA 证书链（自签名证书，证书链为空）
	chainPEM = ""

	return certPEM, chainPEM, nil
}

// GetTLSVersion 获取 TLS 版本
func GetTLSVersion() uint16 {
	return tls.VersionTLS12
}
