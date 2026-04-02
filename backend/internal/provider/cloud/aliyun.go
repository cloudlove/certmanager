package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	cdn "github.com/alibabacloud-go/cdn-20180510/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dcdn "github.com/alibabacloud-go/dcdn-20180115/v3/client"
	slb "github.com/alibabacloud-go/slb-20140515/v4/client"
	"github.com/alibabacloud-go/tea/tea"
)

// AliyunProvider 阿里云部署 Provider
type AliyunProvider struct {
	accessKey  string
	secretKey  string
	region     string
	cdnClient  *cdn.Client
	dcdnClient *dcdn.Client
	slbClient  *slb.Client
}

// NewAliyunProvider 创建阿里云 Provider
func NewAliyunProvider(accessKey, secretKey string) *AliyunProvider {
	provider := &AliyunProvider{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    "cn-hangzhou",
	}

	// 初始化客户端
	if client, err := provider.createCDNClient(); err == nil {
		provider.cdnClient = client
	} else {
		log.Printf("[WARN] Failed to create Aliyun CDN client: %v", err)
	}

	if client, err := provider.createDCDNClient(); err == nil {
		provider.dcdnClient = client
	} else {
		log.Printf("[WARN] Failed to create Aliyun DCDN client: %v", err)
	}

	if client, err := provider.createSLBClient(); err == nil {
		provider.slbClient = client
	} else {
		log.Printf("[WARN] Failed to create Aliyun SLB client: %v", err)
	}

	return provider
}

// createCDNClient 创建 CDN 客户端
func (p *AliyunProvider) createCDNClient() (*cdn.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(p.accessKey),
		AccessKeySecret: tea.String(p.secretKey),
		Endpoint:        tea.String("cdn.aliyuncs.com"),
	}
	return cdn.NewClient(config)
}

// createDCDNClient 创建 DCDN 客户端
func (p *AliyunProvider) createDCDNClient() (*dcdn.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(p.accessKey),
		AccessKeySecret: tea.String(p.secretKey),
		Endpoint:        tea.String("dcdn.aliyuncs.com"),
	}
	return dcdn.NewClient(config)
}

// createSLBClient 创建 SLB 客户端
func (p *AliyunProvider) createSLBClient() (*slb.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(p.accessKey),
		AccessKeySecret: tea.String(p.secretKey),
		Endpoint:        tea.String("slb.aliyuncs.com"),
	}
	return slb.NewClient(config)
}

// getCDNClient 获取 CDN 客户端
func (p *AliyunProvider) getCDNClient() (*cdn.Client, error) {
	if p.cdnClient != nil {
		return p.cdnClient, nil
	}
	return p.createCDNClient()
}

// getDCDNClient 获取 DCDN 客户端
func (p *AliyunProvider) getDCDNClient() (*dcdn.Client, error) {
	if p.dcdnClient != nil {
		return p.dcdnClient, nil
	}
	return p.createDCDNClient()
}

// getSLBClient 获取 SLB 客户端
func (p *AliyunProvider) getSLBClient() (*slb.Client, error) {
	if p.slbClient != nil {
		return p.slbClient, nil
	}
	return p.createSLBClient()
}

// CertSnapshot 证书快照信息
type CertSnapshot struct {
	CertID      string `json:"certId"`
	CertName    string `json:"certName"`
	CertDomain  string `json:"certDomain"`
	SSLProtocol string `json:"sslProtocol"`
}

// DeployCert 部署证书到阿里云资源 (CDN/SLB/DCDN)
func (p *AliyunProvider) DeployCert(ctx context.Context, req DeployRequest) (*DeployResponse, error) {
	log.Printf("[INFO] Deploying certificate to Aliyun: resourceType=%s, resourceID=%s", req.ResourceType, req.ResourceID)

	switch strings.ToLower(req.ResourceType) {
	case "cdn":
		return p.deployToCDN(ctx, req)
	case "slb":
		return p.deployToSLB(ctx, req)
	case "dcdn":
		return p.deployToDCDN(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", req.ResourceType)
	}
}

// deployToCDN 部署证书到 CDN
func (p *AliyunProvider) deployToCDN(ctx context.Context, req DeployRequest) (*DeployResponse, error) {
	client, err := p.getCDNClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get CDN client: %w", err)
	}

	// 生成证书名称
	certName := fmt.Sprintf("certmanager-cdn-%d", time.Now().Unix())

	// 调用 SetCdnDomainSSLCertificate 接口
	setCertReq := &cdn.SetCdnDomainSSLCertificateRequest{
		DomainName:  tea.String(req.ResourceID),
		SSLPub:      tea.String(req.CertPEM),
		SSLProtocol: tea.String("on"),
		SSLPri:      tea.String(req.PrivateKeyPEM),
		CertName:    tea.String(certName),
		CertType:    tea.String("upload"),
	}

	// 如果有证书链，设置 SSLPri 为私钥
	if req.ChainPEM != "" {
		// CDN 证书链通常包含在证书内容中
		combinedCert := req.CertPEM
		if req.ChainPEM != "" {
			combinedCert = combinedCert + "\n" + req.ChainPEM
		}
		setCertReq.SSLPub = tea.String(combinedCert)
	}

	log.Printf("[INFO] Setting CDN domain SSL certificate: domain=%s, certName=%s", req.ResourceID, certName)

	resp, err := client.SetCdnDomainSSLCertificate(setCertReq)
	if err != nil {
		return nil, fmt.Errorf("failed to set CDN SSL certificate: %w", err)
	}

	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("empty response from Aliyun CDN")
	}

	taskID := fmt.Sprintf("aliyun-cdn-%s-%d", req.ResourceID, time.Now().Unix())
	log.Printf("[INFO] CDN SSL certificate set successfully: requestId=%s", tea.StringValue(resp.Body.RequestId))

	return &DeployResponse{
		TaskID:  taskID,
		Status:  "success",
		Message: fmt.Sprintf("Certificate deployed to Aliyun CDN domain %s successfully", req.ResourceID),
	}, nil
}

// deployToDCDN 部署证书到 DCDN (全站加速)
func (p *AliyunProvider) deployToDCDN(ctx context.Context, req DeployRequest) (*DeployResponse, error) {
	client, err := p.getDCDNClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get DCDN client: %w", err)
	}

	// 生成证书名称
	certName := fmt.Sprintf("certmanager-dcdn-%d", time.Now().Unix())

	// 合并证书和证书链
	certPEM := req.CertPEM
	if req.ChainPEM != "" {
		certPEM = certPEM + "\n" + req.ChainPEM
	}

	// 调用 SetDcdnDomainSSLCertificate 接口
	setCertReq := &dcdn.SetDcdnDomainSSLCertificateRequest{
		DomainName:  tea.String(req.ResourceID),
		SSLPub:      tea.String(certPEM),
		SSLProtocol: tea.String("on"),
		SSLPri:      tea.String(req.PrivateKeyPEM),
		CertName:    tea.String(certName),
		CertType:    tea.String("upload"),
	}

	log.Printf("[INFO] Setting DCDN domain SSL certificate: domain=%s, certName=%s", req.ResourceID, certName)

	resp, err := client.SetDcdnDomainSSLCertificate(setCertReq)
	if err != nil {
		return nil, fmt.Errorf("failed to set DCDN SSL certificate: %w", err)
	}

	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("empty response from Aliyun DCDN")
	}

	taskID := fmt.Sprintf("aliyun-dcdn-%s-%d", req.ResourceID, time.Now().Unix())
	log.Printf("[INFO] DCDN SSL certificate set successfully: requestId=%s", tea.StringValue(resp.Body.RequestId))

	return &DeployResponse{
		TaskID:  taskID,
		Status:  "success",
		Message: fmt.Sprintf("Certificate deployed to Aliyun DCDN domain %s successfully", req.ResourceID),
	}, nil
}

// deployToSLB 部署证书到 SLB
func (p *AliyunProvider) deployToSLB(ctx context.Context, req DeployRequest) (*DeployResponse, error) {
	client, err := p.getSLBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get SLB client: %w", err)
	}

	// ResourceID 格式: loadBalancerId:listenerPort
	// 例如: lb-bp123456:443
	parts := strings.Split(req.ResourceID, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid SLB resource ID format, expected 'loadBalancerId:listenerPort', got: %s", req.ResourceID)
	}

	loadBalancerID := parts[0]
	listenerPort := parts[1]

	// 1. 上传服务器证书
	certID, certName, err := p.uploadSLBCertificate(ctx, client, req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload SLB certificate: %w", err)
	}

	log.Printf("[INFO] SLB certificate uploaded: certId=%s, certName=%s", certID, certName)

	// 2. 设置 HTTPS 监听器证书
	err = p.setSLBHTTPSListener(client, loadBalancerID, listenerPort, certID)
	if err != nil {
		return nil, fmt.Errorf("failed to set SLB HTTPS listener: %w", err)
	}

	taskID := fmt.Sprintf("aliyun-slb-%s-%d", req.ResourceID, time.Now().Unix())
	log.Printf("[INFO] SLB certificate deployed successfully: loadBalancerId=%s, port=%s", loadBalancerID, listenerPort)

	return &DeployResponse{
		TaskID:  taskID,
		Status:  "success",
		Message: fmt.Sprintf("Certificate deployed to Aliyun SLB %s port %s successfully", loadBalancerID, listenerPort),
	}, nil
}

// uploadSLBCertificate 上传 SLB 服务器证书
func (p *AliyunProvider) uploadSLBCertificate(ctx context.Context, client *slb.Client, req DeployRequest) (string, string, error) {
	certName := fmt.Sprintf("certmanager-slb-%d", time.Now().Unix())

	uploadReq := &slb.UploadServerCertificateRequest{
		RegionId:              tea.String(p.region),
		ServerCertificate:     tea.String(req.CertPEM),
		PrivateKey:            tea.String(req.PrivateKeyPEM),
		ServerCertificateName: tea.String(certName),
	}

	// 如果有证书链
	if req.ChainPEM != "" {
		uploadReq.ServerCertificate = tea.String(req.CertPEM + "\n" + req.ChainPEM)
	}

	resp, err := client.UploadServerCertificate(uploadReq)
	if err != nil {
		return "", "", err
	}

	if resp == nil || resp.Body == nil {
		return "", "", fmt.Errorf("empty response from SLB upload certificate")
	}

	return tea.StringValue(resp.Body.ServerCertificateId), certName, nil
}

// setSLBHTTPSListener 设置 SLB HTTPS 监听器证书
func (p *AliyunProvider) setSLBHTTPSListener(client *slb.Client, loadBalancerID, listenerPort, certID string) error {
	port := int32(443)
	if listenerPort != "" {
		fmt.Sscanf(listenerPort, "%d", &port)
	}

	setReq := &slb.SetLoadBalancerHTTPSListenerAttributeRequest{
		LoadBalancerId:      tea.String(loadBalancerID),
		ListenerPort:        tea.Int32(port),
		ServerCertificateId: tea.String(certID),
	}

	_, err := client.SetLoadBalancerHTTPSListenerAttribute(setReq)
	return err
}

// GetDeployStatus 获取部署状态
func (p *AliyunProvider) GetDeployStatus(ctx context.Context, taskID string) (string, error) {
	// 阿里云 CDN/SLB/DCDN 的证书部署是同步操作
	// 如果 DeployCert 返回成功，状态就是 success
	if strings.HasPrefix(taskID, "aliyun-cdn-") ||
		strings.HasPrefix(taskID, "aliyun-slb-") ||
		strings.HasPrefix(taskID, "aliyun-dcdn-") {
		return "success", nil
	}
	return "unknown", fmt.Errorf("invalid task ID format: %s", taskID)
}

// GetCurrentCert 获取当前资源上的证书信息
func (p *AliyunProvider) GetCurrentCert(ctx context.Context, resourceID, resourceType string) (string, error) {
	switch strings.ToLower(resourceType) {
	case "cdn":
		return p.getCurrentCDNCert(ctx, resourceID)
	case "slb":
		return p.getCurrentSLBCert(ctx, resourceID)
	case "dcdn":
		return p.getCurrentDCDNCert(ctx, resourceID)
	default:
		return "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// getCurrentCDNCert 获取 CDN 域名当前证书信息
func (p *AliyunProvider) getCurrentCDNCert(ctx context.Context, domain string) (string, error) {
	client, err := p.getCDNClient()
	if err != nil {
		return "", fmt.Errorf("failed to get CDN client: %w", err)
	}

	// 通过 DescribeDomainCertificateInfo 获取证书详情
	certInfoReq := &cdn.DescribeDomainCertificateInfoRequest{
		DomainName: tea.String(domain),
	}
	certInfoResp, err := client.DescribeDomainCertificateInfo(certInfoReq)
	if err != nil {
		return "", fmt.Errorf("failed to describe CDN domain certificate: %w", err)
	}

	if certInfoResp == nil || certInfoResp.Body == nil ||
		certInfoResp.Body.CertInfos == nil || certInfoResp.Body.CertInfos.CertInfo == nil {
		return "", nil
	}

	snapshot := &CertSnapshot{
		SSLProtocol: "off",
	}

	for _, info := range certInfoResp.Body.CertInfos.CertInfo {
		if info != nil {
			snapshot.CertName = tea.StringValue(info.CertName)
			snapshot.CertDomain = tea.StringValue(info.CertDomainName)
			snapshot.CertID = tea.StringValue(info.CertId)
			snapshot.SSLProtocol = tea.StringValue(info.ServerCertificateStatus)
			break
		}
	}

	// 序列化为 JSON
	data, err := json.Marshal(snapshot)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cert snapshot: %w", err)
	}

	return string(data), nil
}

// getCurrentDCDNCert 获取 DCDN 域名当前证书信息
func (p *AliyunProvider) getCurrentDCDNCert(ctx context.Context, domain string) (string, error) {
	client, err := p.getDCDNClient()
	if err != nil {
		return "", fmt.Errorf("failed to get DCDN client: %w", err)
	}

	// 调用 DescribeDcdnDomainDetail 获取域名配置
	req := &dcdn.DescribeDcdnDomainDetailRequest{
		DomainName: tea.String(domain),
	}

	resp, err := client.DescribeDcdnDomainDetail(req)
	if err != nil {
		return "", fmt.Errorf("failed to describe DCDN domain: %w", err)
	}

	if resp == nil || resp.Body == nil || resp.Body.DomainDetail == nil {
		return "", nil
	}

	detail := resp.Body.DomainDetail
	snapshot := &CertSnapshot{
		SSLProtocol: "off",
	}

	// 如果启用 HTTPS
	if detail.SSLProtocol != nil && tea.StringValue(detail.SSLProtocol) == "on" {
		snapshot.SSLProtocol = "on"
		// 获取证书信息
		certInfoReq := &dcdn.DescribeDcdnDomainCertificateInfoRequest{
			DomainName: tea.String(domain),
		}
		certInfoResp, err := client.DescribeDcdnDomainCertificateInfo(certInfoReq)
		if err == nil && certInfoResp != nil && certInfoResp.Body != nil &&
			certInfoResp.Body.CertInfos != nil && certInfoResp.Body.CertInfos.CertInfo != nil {
			for _, info := range certInfoResp.Body.CertInfos.CertInfo {
				if info != nil {
					snapshot.CertName = tea.StringValue(info.CertName)
					snapshot.CertDomain = tea.StringValue(info.CertDomainName)
					snapshot.CertID = tea.StringValue(info.CertId)
					break
				}
			}
		}
	}

	// 序列化为 JSON
	data, err := json.Marshal(snapshot)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cert snapshot: %w", err)
	}

	return string(data), nil
}

// getCurrentSLBCert 获取 SLB 监听器当前证书信息
func (p *AliyunProvider) getCurrentSLBCert(ctx context.Context, resourceID string) (string, error) {
	client, err := p.getSLBClient()
	if err != nil {
		return "", fmt.Errorf("failed to get SLB client: %w", err)
	}

	// 解析 ResourceID
	parts := strings.Split(resourceID, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid SLB resource ID format: %s", resourceID)
	}

	loadBalancerID := parts[0]
	listenerPort := int32(443)
	fmt.Sscanf(parts[1], "%d", &listenerPort)

	// 调用 DescribeLoadBalancerHTTPSListenerAttribute
	req := &slb.DescribeLoadBalancerHTTPSListenerAttributeRequest{
		LoadBalancerId: tea.String(loadBalancerID),
		ListenerPort:   tea.Int32(listenerPort),
	}

	resp, err := client.DescribeLoadBalancerHTTPSListenerAttribute(req)
	if err != nil {
		return "", fmt.Errorf("failed to describe SLB HTTPS listener: %w", err)
	}

	if resp == nil || resp.Body == nil {
		return "", nil
	}

	snapshot := &CertSnapshot{
		SSLProtocol: "on",
		CertID:      tea.StringValue(resp.Body.ServerCertificateId),
	}

	// 序列化为 JSON
	data, err := json.Marshal(snapshot)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cert snapshot: %w", err)
	}

	return string(data), nil
}

// Rollback 回滚证书
func (p *AliyunProvider) Rollback(ctx context.Context, req RollbackRequest) error {
	log.Printf("[INFO] Rolling back certificate: resourceType=%s, resourceID=%s", req.ResourceType, req.ResourceID)

	// 解析旧证书信息
	var snapshot CertSnapshot
	if err := json.Unmarshal([]byte(req.OldCertInfo), &snapshot); err != nil {
		return fmt.Errorf("invalid old cert info: %w", err)
	}

	switch strings.ToLower(req.ResourceType) {
	case "cdn":
		return p.rollbackCDN(ctx, req.ResourceID, &snapshot)
	case "slb":
		return p.rollbackSLB(ctx, req.ResourceID, &snapshot)
	case "dcdn":
		return p.rollbackDCDN(ctx, req.ResourceID, &snapshot)
	default:
		return fmt.Errorf("unsupported resource type: %s", req.ResourceType)
	}
}

// rollbackCDN 回滚 CDN 证书
func (p *AliyunProvider) rollbackCDN(ctx context.Context, domain string, snapshot *CertSnapshot) error {
	// 如果之前没有启用 HTTPS，关闭 HTTPS
	if snapshot.SSLProtocol == "off" {
		client, err := p.getCDNClient()
		if err != nil {
			return fmt.Errorf("failed to get CDN client: %w", err)
		}

		req := &cdn.SetCdnDomainSSLCertificateRequest{
			DomainName:  tea.String(domain),
			SSLProtocol: tea.String("off"),
		}

		_, err = client.SetCdnDomainSSLCertificate(req)
		return err
	}

	// 如果有旧的证书标识，重新设置
	// 注意：阿里云 CDN 的证书回滚需要重新上传证书或使用已有的证书
	// 这里 OldCertInfo 包含的是证书信息，但实际回滚需要重新上传证书内容
	// 由于我们没有旧证书的私钥，这里只能通过证书标识来回滚（如果阿里云支持的话）
	// 实际生产环境中，应该在部署前保存完整的证书信息
	log.Printf("[WARN] CDN rollback requires re-uploading the old certificate. Old cert info: %s", snapshot.CertID)
	return fmt.Errorf("CDN rollback requires the original certificate content which is not available")
}

// rollbackDCDN 回滚 DCDN 证书
func (p *AliyunProvider) rollbackDCDN(ctx context.Context, domain string, snapshot *CertSnapshot) error {
	// 类似 CDN 回滚
	if snapshot.SSLProtocol == "off" {
		client, err := p.getDCDNClient()
		if err != nil {
			return fmt.Errorf("failed to get DCDN client: %w", err)
		}

		req := &dcdn.SetDcdnDomainSSLCertificateRequest{
			DomainName:  tea.String(domain),
			SSLProtocol: tea.String("off"),
		}

		_, err = client.SetDcdnDomainSSLCertificate(req)
		return err
	}

	log.Printf("[WARN] DCDN rollback requires re-uploading the old certificate. Old cert info: %s", snapshot.CertID)
	return fmt.Errorf("DCDN rollback requires the original certificate content which is not available")
}

// rollbackSLB 回滚 SLB 证书
func (p *AliyunProvider) rollbackSLB(ctx context.Context, resourceID string, snapshot *CertSnapshot) error {
	client, err := p.getSLBClient()
	if err != nil {
		return fmt.Errorf("failed to get SLB client: %w", err)
	}

	// 解析 ResourceID
	parts := strings.Split(resourceID, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid SLB resource ID format: %s", resourceID)
	}

	loadBalancerID := parts[0]
	listenerPort := int32(443)
	fmt.Sscanf(parts[1], "%d", &listenerPort)

	if snapshot.CertID == "" {
		return fmt.Errorf("no old certificate ID found for SLB rollback")
	}

	// 使用旧证书 ID 设置监听器
	req := &slb.SetLoadBalancerHTTPSListenerAttributeRequest{
		LoadBalancerId:      tea.String(loadBalancerID),
		ListenerPort:        tea.Int32(listenerPort),
		ServerCertificateId: tea.String(snapshot.CertID),
	}

	_, err = client.SetLoadBalancerHTTPSListenerAttribute(req)
	if err != nil {
		return fmt.Errorf("failed to rollback SLB certificate: %w", err)
	}

	log.Printf("[INFO] SLB certificate rolled back successfully: loadBalancerId=%s, port=%d, certId=%s",
		loadBalancerID, listenerPort, snapshot.CertID)
	return nil
}
