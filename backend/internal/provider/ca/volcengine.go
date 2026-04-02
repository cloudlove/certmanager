package ca

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
	"sort"
	"strings"
	"time"
)

// VolcengineProvider 火山云 CA Provider
type VolcengineProvider struct {
	accessKey  string
	secretKey  string
	region     string
	httpClient *http.Client
}

// 火山云证书服务 API 端点
const (
	volcengineCertServiceEndpoint = "open.volcengineapi.com"
	volcengineCertServiceHost     = "certificate_service"
	volcengineCertServiceVersion  = "2024-10-01"
)

// NewVolcengineProvider 创建火山云 Provider 实例
func NewVolcengineProvider(accessKey, secretKey string) *VolcengineProvider {
	return &VolcengineProvider{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    "cn-beijing", // 默认区域
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// VolcengineAPIResponse 通用 API 响应
type VolcengineAPIResponse struct {
	ResponseMetadata struct {
		RequestId string `json:"RequestId"`
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

// QuickApplyCertificateResult 快速申请证书结果
type QuickApplyCertificateResult struct {
	InstanceId   string `json:"InstanceId"`
	OrderId      string `json:"OrderId"`
	InstanceName string `json:"InstanceName"`
	Status       string `json:"Status"`
}

// GetInstanceResult 获取证书实例结果
type GetInstanceResult struct {
	InstanceId    string `json:"InstanceId"`
	InstanceName  string `json:"InstanceName"`
	InstanceIdStr string `json:"InstanceIdStr"`
	Status        string `json:"Status"`
	CommonName    string `json:"CommonName"`
	CertId        string `json:"CertId"`
	CertIdStr     string `json:"CertIdStr"`
	NotAfter      string `json:"NotAfter"`
	NotBefore     string `json:"NotBefore"`
}

// DownloadCertResult 下载证书结果
type DownloadCertResult struct {
	Certificate      string `json:"Certificate"`
	CertificateChain string `json:"CertificateChain"`
	PrivateKey       string `json:"PrivateKey,omitempty"`
}

// ApplyCert 申请证书
// 调用火山云证书中心 API: QuickApplyCertificate
func (p *VolcengineProvider) ApplyCert(ctx context.Context, req ApplyCertRequest) (*ApplyCertResponse, error) {
	log.Printf("[INFO] Applying certificate from Volcengine Certificate Service: domain=%s, productType=%s, validateType=%s", req.Domain, req.ProductType, req.ValidateType)

	// 产品类型映射
	productCode := mapVolcengineProductType(req.ProductType)

	// 构建请求参数
	params := map[string]interface{}{
		"ProductCode": productCode,
		"Domain":      req.Domain,
	}

	// 如果提供了 CSR，添加 CSR 参数
	if req.CSRPEM != "" {
		params["Csr"] = req.CSRPEM
	}

	// 如果指定了有效期
	if req.ValidityYears > 0 {
		// 火山云证书有效期通常由产品类型决定，这里尝试设置
		params["ValidityPeriod"] = fmt.Sprintf("%dY", req.ValidityYears)
	}

	// 添加验证方式（如果火山云API支持）
	if req.ValidateType != "" {
		params["ValidateType"] = req.ValidateType
	}

	// 调用 QuickApplyCertificate API
	result := &QuickApplyCertificateResult{}
	err := p.callAPI(ctx, "QuickApplyCertificate", params, result)
	if err != nil {
		return nil, fmt.Errorf("failed to apply certificate from Volcengine: %w", err)
	}

	log.Printf("[INFO] Certificate applied successfully: instanceId=%s, orderId=%s", result.InstanceId, result.OrderId)

	// 使用 InstanceId 作为 orderID
	orderID := result.InstanceId
	if orderID == "" {
		orderID = result.OrderId
	}

	return &ApplyCertResponse{
		OrderID: orderID,
		Status:  "pending",
	}, nil
}

// GetCertStatus 获取证书状态
// 调用火山云证书中心 API: DescribeInstance 或 GetInstance
func (p *VolcengineProvider) GetCertStatus(ctx context.Context, orderID string) (*CertStatusResponse, error) {
	log.Printf("[INFO] Getting certificate status from Volcengine: orderId=%s", orderID)

	params := map[string]interface{}{
		"InstanceId": orderID,
	}

	result := &GetInstanceResult{}
	err := p.callAPI(ctx, "GetInstance", params, result)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate status: %w", err)
	}

	// 映射火山云状态到内部状态
	status := mapVolcengineCertStatus(result.Status)

	log.Printf("[INFO] Certificate status: orderId=%s, volcStatus=%s, mappedStatus=%s", orderID, result.Status, status)

	// 构建响应
	resp := &CertStatusResponse{
		OrderID: orderID,
		Status:  status,
	}

	// 如果证书已签发，尝试获取证书内容
	if status == "issued" && result.CertId != "" {
		// 调用下载证书API获取证书内容
		downloadParams := map[string]interface{}{
			"InstanceId": orderID,
		}
		downloadResult := &DownloadCertResult{}
		if err := p.callAPI(ctx, "DownloadCertificate", downloadParams, downloadResult); err == nil {
			resp.Certificate = downloadResult.Certificate
			resp.PrivateKey = downloadResult.PrivateKey
		}
	}

	// 注意：火山云API可能不直接提供验证信息字段，保持兼容（字段留空）
	// 如果需要验证信息，可能需要调用其他API或在后续版本中扩展

	return resp, nil
}

// DownloadCert 下载证书
// 调用火山云证书中心 API: DownloadCertificate
func (p *VolcengineProvider) DownloadCert(ctx context.Context, orderID string) (certPEM string, chainPEM string, err error) {
	log.Printf("[INFO] Downloading certificate from Volcengine: orderId=%s", orderID)

	// 首先检查证书状态
	statusResp, err := p.GetCertStatus(ctx, orderID)
	if err != nil {
		return "", "", fmt.Errorf("failed to check certificate status: %w", err)
	}

	if statusResp.Status != "issued" {
		return "", "", fmt.Errorf("certificate is not issued yet, current status: %s", statusResp.Status)
	}

	// 调用下载证书 API
	params := map[string]interface{}{
		"InstanceId": orderID,
	}

	result := &DownloadCertResult{}
	err = p.callAPI(ctx, "DownloadCertificate", params, result)
	if err != nil {
		return "", "", fmt.Errorf("failed to download certificate: %w", err)
	}

	certPEM = result.Certificate
	chainPEM = result.CertificateChain

	if certPEM == "" {
		return "", "", fmt.Errorf("certificate content is empty")
	}

	log.Printf("[INFO] Certificate downloaded successfully: orderId=%s, certLength=%d", orderID, len(certPEM))

	return certPEM, chainPEM, nil
}

// mapVolcengineProductType 映射产品类型到火山云产品代码
func mapVolcengineProductType(productType string) string {
	switch productType {
	case "DV":
		return "free_dv" // 免费 DV 证书
	case "OV":
		return "ov_pro" // OV 专业版
	case "EV":
		return "ev_pro" // EV 专业版
	default:
		return "free_dv" // 默认免费 DV
	}
}

// mapVolcengineCertStatus 映射火山云证书状态到内部状态
func mapVolcengineCertStatus(volcStatus string) string {
	switch volcStatus {
	case "issued", "deployed", "active":
		return "issued"
	case "failed", "error", "expired", "revoked":
		return "failed"
	case "domain_verify":
		return "domain_verify"
	case "process", "processing":
		return "process"
	case "pending", "checking", "verifying", "applying":
		return "pending"
	default:
		return "pending"
	}
}

// callAPI 调用火山云 OpenAPI
func (p *VolcengineProvider) callAPI(ctx context.Context, action string, params map[string]interface{}, result interface{}) error {
	// 构建 URL
	scheme := "https"
	host := volcengineCertServiceEndpoint
	path := "/"

	// 构建查询参数
	query := url.Values{}
	query.Set("Action", action)
	query.Set("Version", volcengineCertServiceVersion)

	// 构建 URL
	requestURL := fmt.Sprintf("%s://%s%s?%s", scheme, host, path, query.Encode())

	// 构建请求体
	body, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
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
	signature := p.signRequest(req, string(body))
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
	if apiResp.Result != nil {
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
func (p *VolcengineProvider) signRequest(req *http.Request, body string) string {
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
	kService := hmacSHA256(kRegion, volcengineCertServiceHost)
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

// sortQueryString 对查询参数进行排序
func sortQueryString(query url.Values) string {
	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(url.QueryEscape(k))
		buf.WriteByte('=')
		buf.WriteString(url.QueryEscape(query.Get(k)))
	}
	return buf.String()
}
