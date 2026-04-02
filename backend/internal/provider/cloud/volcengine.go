package cloud

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// VolcengineProvider 火山引擎部署 Provider
type VolcengineProvider struct {
	accessKey  string
	secretKey  string
	region     string
	httpClient *http.Client
}

// 火山云服务端点
const (
	volcengineOpenAPIEndpoint = "open.volcengineapi.com"
	volcengineCDNService      = "CDN"
	volcengineCLBService      = "clb"
)

// NewVolcengineProvider 创建火山引擎 Provider
func NewVolcengineProvider(accessKey, secretKey string) *VolcengineProvider {
	return &VolcengineProvider{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    "cn-beijing",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// VolcengineAPIResponse 通用 API 响应
type VolcengineAPIResponse struct {
	ResponseMetadata struct {
		RequestID string `json:"RequestId"`
		Action    string `json:"Action"`
		Version   string `json:"Version"`
		Service   string `json:"Service"`
		Region    string `json:"Region"`
		Error     *struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error,omitempty"`
	} `json:"ResponseMetadata"`
	Result interface{} `json:"Result"`
}

// CDN 相关响应结构
type volcengineAddCertificateResult struct {
	CertID   string `json:"CertId"`
	CertName string `json:"CertName"`
}

type volcengineDescribeDomainConfigResult struct {
	Domain string `json:"Domain"`
	HTTPS  *struct {
		EnableHTTPS        bool   `json:"EnableHTTPS"`
		CertID             string `json:"CertId"`
		CertDomain         string `json:"CertDomain"`
		HTTP2Enable        bool   `json:"HTTP2Enable"`
		ForceRedirectHTTPS bool   `json:"ForceRedirectHTTPS"`
	} `json:"HTTPS"`
}

// CLB 相关响应结构
type volcengineUploadCertificateResult struct {
	CertificateID   string `json:"CertificateId"`
	CertificateName string `json:"CertificateName"`
	ExpireTime      string `json:"ExpireTime"`
}

type volcengineDescribeListenerAttributesResult struct {
	ListenerID    string `json:"ListenerId"`
	Protocol      string `json:"Protocol"`
	Port          int    `json:"Port"`
	CertificateID string `json:"CertificateId"`
}

// DeployCert 部署证书到火山引擎资源 (CDN/CLB)
func (p *VolcengineProvider) DeployCert(ctx context.Context, req DeployRequest) (*DeployResponse, error) {
	log.Printf("[INFO] Deploying certificate to Volcengine: resourceType=%s, resourceID=%s", req.ResourceType, req.ResourceID)

	switch strings.ToLower(req.ResourceType) {
	case "cdn":
		return p.deployToCDN(ctx, req)
	case "clb":
		return p.deployToCLB(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", req.ResourceType)
	}
}

// deployToCDN 部署证书到 CDN
func (p *VolcengineProvider) deployToCDN(ctx context.Context, req DeployRequest) (*DeployResponse, error) {
	// 1. 上传证书到火山云 CDN 证书中心
	certID, err := p.uploadCDNCertificate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload CDN certificate: %w", err)
	}

	log.Printf("[INFO] Certificate uploaded to CDN certificate center: certId=%s", certID)

	// 2. 配置域名 HTTPS 绑定证书
	err = p.bindCDNCertificate(ctx, req.ResourceID, certID)
	if err != nil {
		return nil, fmt.Errorf("failed to bind certificate to CDN domain: %w", err)
	}

	taskID := fmt.Sprintf("volcengine-cdn-%s-%d", req.ResourceID, time.Now().Unix())
	log.Printf("[INFO] Certificate deployed to Volcengine CDN successfully: domain=%s, certId=%s", req.ResourceID, certID)

	return &DeployResponse{
		TaskID:  taskID,
		Status:  "success",
		Message: fmt.Sprintf("Certificate deployed to Volcengine CDN domain %s successfully", req.ResourceID),
	}, nil
}

// uploadCDNCertificate 上传证书到 CDN 证书中心
func (p *VolcengineProvider) uploadCDNCertificate(ctx context.Context, req DeployRequest) (string, error) {
	// 合并证书链
	certChain := req.CertPEM
	if req.ChainPEM != "" {
		certChain = certChain + "\n" + req.ChainPEM
	}

	// 生成证书名称
	certName := fmt.Sprintf("certmanager-%d", time.Now().Unix())

	params := map[string]interface{}{
		"Certificate": certChain,
		"PrivateKey":  req.PrivateKeyPEM,
		"Source":      "volc_cert_center",
		"Desc":        certName,
	}

	result := &volcengineAddCertificateResult{}
	err := p.callCDNAPI(ctx, "AddCertificate", params, result)
	if err != nil {
		return "", err
	}

	return result.CertID, nil
}

// bindCDNCertificate 绑定证书到 CDN 域名
func (p *VolcengineProvider) bindCDNCertificate(ctx context.Context, domain, certID string) error {
	params := map[string]interface{}{
		"Domain": domain,
		"HTTPS": map[string]interface{}{
			"EnableHTTPS": true,
			"CertId":      certID,
			"HTTP2Enable": true,
		},
	}

	err := p.callCDNAPI(ctx, "UpdateHttps", params, nil)
	if err != nil {
		return err
	}

	return nil
}

// deployToCLB 部署证书到 CLB
func (p *VolcengineProvider) deployToCLB(ctx context.Context, req DeployRequest) (*DeployResponse, error) {
	// ResourceID 格式: listenerId (监听器ID)
	listenerID := req.ResourceID

	// 1. 上传证书到 CLB 证书中心
	certID, err := p.uploadCLBCertificate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload CLB certificate: %w", err)
	}

	log.Printf("[INFO] Certificate uploaded to CLB certificate center: certId=%s", certID)

	// 2. 修改监听器绑定新证书
	err = p.modifyListenerCertificate(ctx, listenerID, certID)
	if err != nil {
		return nil, fmt.Errorf("failed to modify listener certificate: %w", err)
	}

	taskID := fmt.Sprintf("volcengine-clb-%s-%d", listenerID, time.Now().Unix())
	log.Printf("[INFO] Certificate deployed to Volcengine CLB successfully: listenerId=%s, certId=%s", listenerID, certID)

	return &DeployResponse{
		TaskID:  taskID,
		Status:  "success",
		Message: fmt.Sprintf("Certificate deployed to Volcengine CLB listener %s successfully", listenerID),
	}, nil
}

// uploadCLBCertificate 上传证书到 CLB 证书中心
func (p *VolcengineProvider) uploadCLBCertificate(ctx context.Context, req DeployRequest) (string, error) {
	certName := fmt.Sprintf("certmanager-clb-%d", time.Now().Unix())

	params := map[string]interface{}{
		"CertificateName":  certName,
		"PrivateKey":       req.PrivateKeyPEM,
		"Certificate":      req.CertPEM,
		"CertificateChain": req.ChainPEM,
	}

	result := &volcengineUploadCertificateResult{}
	err := p.callCLBAPI(ctx, "UploadCertificate", params, result)
	if err != nil {
		return "", err
	}

	return result.CertificateID, nil
}

// modifyListenerCertificate 修改监听器证书
func (p *VolcengineProvider) modifyListenerCertificate(ctx context.Context, listenerID, certID string) error {
	params := map[string]interface{}{
		"ListenerId":    listenerID,
		"CertificateId": certID,
	}

	err := p.callCLBAPI(ctx, "ModifyListenerAttributes", params, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetDeployStatus 获取部署状态
func (p *VolcengineProvider) GetDeployStatus(ctx context.Context, taskID string) (string, error) {
	// 火山云 CDN/CLB 的证书部署是同步操作，如果 DeployCert 返回成功，状态就是 success
	// 对于异步任务，可以通过任务 ID 解析资源类型查询实际状态
	if strings.HasPrefix(taskID, "volcengine-cdn-") || strings.HasPrefix(taskID, "volcengine-clb-") {
		return "success", nil
	}
	return "unknown", fmt.Errorf("invalid task ID format: %s", taskID)
}

// GetCurrentCert 获取当前资源上的证书信息
func (p *VolcengineProvider) GetCurrentCert(ctx context.Context, resourceID, resourceType string) (string, error) {
	switch strings.ToLower(resourceType) {
	case "cdn":
		return p.getCurrentCDNCert(ctx, resourceID)
	case "clb":
		return p.getCurrentCLBCert(ctx, resourceID)
	default:
		return "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// getCurrentCDNCert 获取 CDN 域名当前绑定的证书信息
func (p *VolcengineProvider) getCurrentCDNCert(ctx context.Context, domain string) (string, error) {
	params := map[string]interface{}{
		"Domain": domain,
	}

	result := &volcengineDescribeDomainConfigResult{}
	err := p.callCDNAPI(ctx, "DescribeDomainDetail", params, result)
	if err != nil {
		return "", fmt.Errorf("failed to get CDN domain config: %w", err)
	}

	if result.HTTPS == nil || !result.HTTPS.EnableHTTPS {
		return "", nil // 未启用 HTTPS，无证书
	}

	// 返回证书 ID 作为旧证书信息（用于回滚）
	return fmt.Sprintf("certId:%s,domain:%s", result.HTTPS.CertID, result.HTTPS.CertDomain), nil
}

// getCurrentCLBCert 获取 CLB 监听器当前绑定的证书信息
func (p *VolcengineProvider) getCurrentCLBCert(ctx context.Context, listenerID string) (string, error) {
	params := map[string]interface{}{
		"ListenerId": listenerID,
	}

	result := &volcengineDescribeListenerAttributesResult{}
	err := p.callCLBAPI(ctx, "DescribeListenerAttributes", params, result)
	if err != nil {
		return "", fmt.Errorf("failed to get CLB listener config: %w", err)
	}

	if result.CertificateID == "" {
		return "", nil // 无证书绑定
	}

	return fmt.Sprintf("certId:%s", result.CertificateID), nil
}

// Rollback 回滚证书
func (p *VolcengineProvider) Rollback(ctx context.Context, req RollbackRequest) error {
	log.Printf("[INFO] Rolling back certificate: resourceType=%s, resourceID=%s", req.ResourceType, req.ResourceID)

	switch strings.ToLower(req.ResourceType) {
	case "cdn":
		return p.rollbackCDN(ctx, req)
	case "clb":
		return p.rollbackCLB(ctx, req)
	default:
		return fmt.Errorf("unsupported resource type: %s", req.ResourceType)
	}
}

// rollbackCDN 回滚 CDN 证书
func (p *VolcengineProvider) rollbackCDN(ctx context.Context, req RollbackRequest) error {
	// OldCertInfo 格式: certId:xxx,domain:xxx
	oldCertID := ""
	if strings.HasPrefix(req.OldCertInfo, "certId:") {
		parts := strings.Split(req.OldCertInfo, ",")
		if len(parts) > 0 {
			oldCertID = strings.TrimPrefix(parts[0], "certId:")
		}
	}

	if oldCertID == "" {
		return fmt.Errorf("invalid old cert info for CDN rollback: %s", req.OldCertInfo)
	}

	// 绑定旧证书
	err := p.bindCDNCertificate(ctx, req.ResourceID, oldCertID)
	if err != nil {
		return fmt.Errorf("failed to rollback CDN certificate: %w", err)
	}

	log.Printf("[INFO] CDN certificate rolled back successfully: domain=%s, oldCertId=%s", req.ResourceID, oldCertID)
	return nil
}

// rollbackCLB 回滚 CLB 证书
func (p *VolcengineProvider) rollbackCLB(ctx context.Context, req RollbackRequest) error {
	// OldCertInfo 格式: certId:xxx
	oldCertID := ""
	if strings.HasPrefix(req.OldCertInfo, "certId:") {
		oldCertID = strings.TrimPrefix(req.OldCertInfo, "certId:")
	}

	if oldCertID == "" {
		return fmt.Errorf("invalid old cert info for CLB rollback: %s", req.OldCertInfo)
	}

	// 修改监听器绑定旧证书
	err := p.modifyListenerCertificate(ctx, req.ResourceID, oldCertID)
	if err != nil {
		return fmt.Errorf("failed to rollback CLB certificate: %w", err)
	}

	log.Printf("[INFO] CLB certificate rolled back successfully: listenerId=%s, oldCertId=%s", req.ResourceID, oldCertID)
	return nil
}

// callCDNAPI 调用火山云 CDN OpenAPI
func (p *VolcengineProvider) callCDNAPI(ctx context.Context, action string, params map[string]interface{}, result interface{}) error {
	return p.callAPI(ctx, volcengineCDNService, action, params, result)
}

// callCLBAPI 调用火山云 CLB OpenAPI
func (p *VolcengineProvider) callCLBAPI(ctx context.Context, action string, params map[string]interface{}, result interface{}) error {
	return p.callAPI(ctx, volcengineCLBService, action, params, result)
}

// callAPI 调用火山云 OpenAPI
func (p *VolcengineProvider) callAPI(ctx context.Context, service, action string, params map[string]interface{}, result interface{}) error {
	scheme := "https"
	host := volcengineOpenAPIEndpoint
	path := "/"

	// 构建查询参数
	query := url.Values{}
	query.Set("Action", action)
	query.Set("Version", "2022-02-01")

	// 构建 URL
	requestURL := fmt.Sprintf("%s://%s%s?%s", scheme, host, path, query.Encode())

	// 构建请求体
	var body []byte
	var err error
	if params != nil {
		body, err = json.Marshal(params)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", host)

	// 添加签名
	now := time.Now().UTC()
	req.Header.Set("X-Date", now.Format("20060102T150405Z"))

	// 计算签名
	signature := p.signRequest(req, string(body), service)
	req.Header.Set("Authorization", signature)

	// 发送请求
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// 解析响应
	var apiResp VolcengineAPIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %w, body: %s", err, string(respBody))
	}

	// 检查错误
	if apiResp.ResponseMetadata.Error != nil {
		return fmt.Errorf("API error: %s - %s",
			apiResp.ResponseMetadata.Error.Code,
			apiResp.ResponseMetadata.Error.Message)
	}

	// 解析结果
	if result != nil && apiResp.Result != nil {
		resultBytes, err := json.Marshal(apiResp.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		if err := json.Unmarshal(resultBytes, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}

// signRequest 对请求进行签名
func (p *VolcengineProvider) signRequest(req *http.Request, body string, service string) string {
	// 获取请求信息
	method := req.Method
	host := req.Host
	if host == "" {
		host = req.Header.Get("Host")
	}
	path := req.URL.Path
	if path == "" {
		path = "/"
	}

	// 获取时间戳
	xDate := req.Header.Get("X-Date")
	shortDate := xDate[:8]

	// 构建 CanonicalRequest
	canonicalURI := path
	canonicalQueryString := req.URL.Query().Encode()
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-content-sha256:%s\nx-date:%s\n",
		"application/json",
		host,
		hex.EncodeToString(sha256Hash([]byte(body))),
		xDate,
	)
	signedHeaders := "content-type;host;x-content-sha256;x-date"

	// 计算 body hash
	bodyHash := hex.EncodeToString(sha256Hash([]byte(body)))

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		method,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		bodyHash,
	)

	// 构建 StringToSign
	algorithm := "HMAC-SHA256"
	credentialScope := fmt.Sprintf("%s/%s/request", shortDate, p.region)
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm,
		xDate,
		credentialScope,
		hex.EncodeToString(sha256Hash([]byte(canonicalRequest))),
	)

	// 计算签名
	kDate := hmacSHA256([]byte(p.secretKey), shortDate)
	kRegion := hmacSHA256(kDate, p.region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "request")
	signature := hmacSHA256Hex(kSigning, stringToSign)

	// 构建 Authorization header
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		p.accessKey,
		credentialScope,
		signedHeaders,
		signature,
	)

	return authorization
}

// sha256Hash 计算 SHA256 哈希
func sha256Hash(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

// hmacSHA256 计算 HMAC-SHA256
func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// hmacSHA256Hex 计算 HMAC-SHA256 并返回十六进制字符串
func hmacSHA256Hex(key []byte, data string) string {
	return hex.EncodeToString(hmacSHA256(key, data))
}
