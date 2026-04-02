package ca

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	cas "github.com/alibabacloud-go/cas-20200407/v3/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

// AliyunProvider 阿里云 CA Provider
type AliyunProvider struct {
	accessKey string
	secretKey string
	client    *cas.Client
}

// NewAliyunProvider 创建阿里云 Provider 实例
func NewAliyunProvider(accessKey, secretKey string) *AliyunProvider {
	provider := &AliyunProvider{
		accessKey: accessKey,
		secretKey: secretKey,
	}
	// 初始化 CAS client
	client, err := provider.createClient()
	if err != nil {
		log.Printf("[WARN] Failed to create Aliyun CAS client: %v, will create on demand", err)
	} else {
		provider.client = client
	}
	return provider
}

// createClient 创建 CAS 客户端
func (p *AliyunProvider) createClient() (*cas.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(p.accessKey),
		AccessKeySecret: tea.String(p.secretKey),
		Endpoint:        tea.String("cas.aliyuncs.com"),
	}
	client, err := cas.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create CAS client: %w", err)
	}
	return client, nil
}

// getClient 获取或创建 CAS 客户端
func (p *AliyunProvider) getClient() (*cas.Client, error) {
	if p.client != nil {
		return p.client, nil
	}
	return p.createClient()
}

// ApplyCert 申请证书
// 调用阿里云 CAS API: CreateCertificateForPackageRequest
func (p *AliyunProvider) ApplyCert(ctx context.Context, req ApplyCertRequest) (*ApplyCertResponse, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get CAS client: %w", err)
	}

	// 产品类型映射: DV/OV/EV -> 阿里云产品代码
	// 参考: https://help.aliyun.com/document_detail/126507.html
	productCode := "digicert-free-1-free" // 默认 Digicert 免费证书
	switch req.ProductType {
	case "DV":
		productCode = "digicert-free-1-free" // Digicert 免费DV证书
	case "OV":
		productCode = "digicert-ov-1-enterprise" // Digicert OV 企业版
	case "EV":
		productCode = "digicert-ev-1-enterprise" // Digicert EV 企业版
	}

	// 构建请求参数
	// CSR 需要转换为 base64 格式
	csrBase64 := base64.StdEncoding.EncodeToString([]byte(req.CSRPEM))

	applyReq := &cas.CreateCertificateForPackageRequestRequest{
		ProductCode:  tea.String(productCode),
		Domain:       tea.String(req.Domain),
		Csr:          tea.String(csrBase64),
		ValidateType: tea.String(req.ValidateType), // 验证方式: DNS 或 FILE
	}

	// 注意: 阿里云证书有效期由产品类型决定，不再支持自定义设置有效期

	log.Printf("[INFO] Applying certificate from Aliyun CAS: domain=%s, product=%s, validateType=%s", req.Domain, productCode, req.ValidateType)

	resp, err := client.CreateCertificateForPackageRequest(applyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to apply certificate from Aliyun: %w", err)
	}

	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("empty response from Aliyun CAS")
	}

	orderID := fmt.Sprintf("%d", tea.Int64Value(resp.Body.OrderId))
	log.Printf("[INFO] Certificate applied successfully: orderId=%s", orderID)

	return &ApplyCertResponse{
		OrderID: orderID,
		Status:  "pending",
	}, nil
}

// GetCertStatus 获取证书状态
// 调用阿里云 CAS API: DescribeCertificateState
func (p *AliyunProvider) GetCertStatus(ctx context.Context, orderID string) (*CertStatusResponse, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get CAS client: %w", err)
	}

	log.Printf("[INFO] Getting certificate status from Aliyun CAS: orderId=%s", orderID)

	// 查询证书状态
	// OrderId 需要转换为 int64
	var orderIdInt int64
	fmt.Sscanf(orderID, "%d", &orderIdInt)
	req := &cas.DescribeCertificateStateRequest{
		OrderId: tea.Int64(orderIdInt),
	}

	resp, err := client.DescribeCertificateState(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate status: %w", err)
	}

	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("empty response from Aliyun CAS")
	}

	// 映射阿里云状态到内部状态
	// 阿里云状态: certificate, domain_verify, process, verify_fail, payed, unknow
	aliStatus := tea.StringValue(resp.Body.Type)
	status := mapAliyunCertStatus(aliStatus)

	log.Printf("[INFO] Certificate status: orderId=%s, aliStatus=%s, mappedStatus=%s", orderID, aliStatus, status)

	// 构建响应
	result := &CertStatusResponse{
		OrderID: orderID,
		Status:  status,
	}

	// 根据状态填充验证信息或证书内容
	switch aliStatus {
	case "domain_verify":
		// 提取DNS验证信息
		result.RecordDomain = tea.StringValue(resp.Body.RecordDomain)
		result.RecordType = tea.StringValue(resp.Body.RecordType)
		result.RecordValue = tea.StringValue(resp.Body.RecordValue)
		// 提取文件验证信息
		result.Uri = tea.StringValue(resp.Body.Uri)
		result.Content = tea.StringValue(resp.Body.Content)
	case "certificate":
		// 提取证书内容
		result.Certificate = tea.StringValue(resp.Body.Certificate)
		result.PrivateKey = tea.StringValue(resp.Body.PrivateKey)
	}

	return result, nil
}

// mapAliyunCertStatus 映射阿里云证书状态到内部状态
func mapAliyunCertStatus(aliStatus string) string {
	switch aliStatus {
	case "certificate":
		return "issued"
	case "verify_fail":
		return "failed"
	case "domain_verify":
		return "domain_verify"
	case "process":
		return "process"
	case "payed", "unknow":
		return "pending"
	default:
		return "pending"
	}
}

// DownloadCert 下载证书
// 调用阿里云 CAS API 获取证书内容，复用 GetCertStatus 逻辑
func (p *AliyunProvider) DownloadCert(ctx context.Context, orderID string) (certPEM string, chainPEM string, err error) {
	log.Printf("[INFO] Downloading certificate from Aliyun CAS: orderId=%s", orderID)

	// 复用 GetCertStatus 获取证书状态和内容
	statusResp, err := p.GetCertStatus(ctx, orderID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get certificate status: %w", err)
	}

	// 检查是否已签发
	if statusResp.Status != "issued" {
		return "", "", fmt.Errorf("certificate is not issued yet, current status: %s", statusResp.Status)
	}

	// 从响应中获取证书内容
	certPEM = statusResp.Certificate

	if certPEM == "" {
		return "", "", fmt.Errorf("certificate content is empty")
	}

	log.Printf("[INFO] Certificate downloaded successfully: orderId=%s, certLength=%d", orderID, len(certPEM))

	return certPEM, "", nil
}

// 注意：以下代码已不再需要，保留函数签名以兼容接口
func (p *AliyunProvider) downloadCertDetail(ctx context.Context, certID string) (certPEM string, chainPEM string, err error) {
	client, err := p.getClient()
	if err != nil {
		return "", "", fmt.Errorf("failed to get CAS client: %w", err)
	}

	// 解析证书 ID
	var certIdInt int64
	fmt.Sscanf(certID, "%d", &certIdInt)

	// 获取证书详情
	detailReq := &cas.GetUserCertificateDetailRequest{
		CertId: tea.Int64(certIdInt),
	}

	detailResp, err := client.GetUserCertificateDetail(detailReq)
	if err != nil {
		return "", "", fmt.Errorf("failed to get certificate detail: %w", err)
	}

	if detailResp == nil || detailResp.Body == nil {
		return "", "", fmt.Errorf("empty detail response from Aliyun CAS")
	}

	// 获取证书内容
	certPEM = tea.StringValue(detailResp.Body.Cert)

	if certPEM == "" {
		return "", "", fmt.Errorf("certificate content is empty")
	}

	return certPEM, "", nil
}
