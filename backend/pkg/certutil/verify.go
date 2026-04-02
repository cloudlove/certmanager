package certutil

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net"
	"strings"
	"time"
)

// CertInfo 证书信息结构体
type CertInfo struct {
	Domain          string    `json:"domain"`
	Issuer          string    `json:"issuer"`
	NotBefore       time.Time `json:"not_before"`
	NotAfter        time.Time `json:"not_after"`
	Fingerprint     string    `json:"fingerprint"`
	SerialNumber    string    `json:"serial_number"`
	SANs            []string  `json:"sans"`
	KeyAlgorithm    string    `json:"key_algorithm"`
	IsExpired       bool      `json:"is_expired"`
	DaysUntilExpiry int       `json:"days_until_expiry"`
}

// VerifyRemoteCert TLS连接远端获取证书信息
func VerifyRemoteCert(domain string) (*CertInfo, error) {
	// 清理域名，去除端口
	host := domain
	if strings.Contains(domain, ":") {
		host, _, _ = net.SplitHostPort(domain)
	}

	// 默认使用443端口
	addr := net.JoinHostPort(host, "443")

	// 建立TLS连接
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	defer conn.Close()

	// 获取证书
	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}

	cert := state.PeerCertificates[0]
	return extractCertInfo(cert, domain), nil
}

// ParseCertificate 解析 PEM 格式证书
func ParseCertificate(certPEM string) (*CertInfo, error) {
	cert, err := parsePEMCertificate(certPEM)
	if err != nil {
		return nil, err
	}

	// 尝试从证书中提取域名
	domain := ""
	if len(cert.DNSNames) > 0 {
		domain = cert.DNSNames[0]
	} else if cert.Subject.CommonName != "" {
		domain = cert.Subject.CommonName
	}

	return extractCertInfo(cert, domain), nil
}

// CheckCertExpiry 检查证书是否过期或即将过期
func CheckCertExpiry(certPEM string, warningDays int) (isExpired bool, daysLeft int, err error) {
	cert, err := parsePEMCertificate(certPEM)
	if err != nil {
		return false, 0, err
	}

	now := time.Now()
	if now.After(cert.NotAfter) {
		return true, 0, nil
	}

	daysLeft = int(cert.NotAfter.Sub(now).Hours() / 24)
	isExpired = daysLeft <= warningDays

	return isExpired, daysLeft, nil
}

// VerifyCertChain 验证证书链完整性
func VerifyCertChain(certPEM, chainPEM string) error {
	cert, err := parsePEMCertificate(certPEM)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// 创建证书池
	pool := x509.NewCertPool()

	// 添加证书链
	if chainPEM != "" {
		block, rest := pem.Decode([]byte(chainPEM))
		for block != nil {
			if block.Type == "CERTIFICATE" {
				pool.AppendCertsFromPEM(pem.EncodeToMemory(block))
			}
			block, rest = pem.Decode(rest)
		}
	}

	// 尝试使用系统根证书池
	roots, err := x509.SystemCertPool()
	if err == nil {
		for _, c := range roots.Subjects() {
			pool.AppendCertsFromPEM(c)
		}
	}

	// 验证证书
	opts := x509.VerifyOptions{
		Roots:         pool,
		Intermediates: x509.NewCertPool(),
		CurrentTime:   time.Now(),
	}

	_, err = cert.Verify(opts)
	if err != nil {
		return fmt.Errorf("certificate chain verification failed: %w", err)
	}

	return nil
}

// MatchDomain 验证证书域名是否匹配
func MatchDomain(certPEM string, domain string) (bool, error) {
	cert, err := parsePEMCertificate(certPEM)
	if err != nil {
		return false, err
	}

	// 检查 Common Name
	if matchesDomain(cert.Subject.CommonName, domain) {
		return true, nil
	}

	// 检查 SANs
	for _, san := range cert.DNSNames {
		if matchesDomain(san, domain) {
			return true, nil
		}
	}

	return false, nil
}

// matchesDomain 检查域名是否匹配（支持通配符）
func matchesDomain(pattern, domain string) bool {
	pattern = strings.ToLower(pattern)
	domain = strings.ToLower(domain)

	// 直接匹配
	if pattern == domain {
		return true
	}

	// 通配符匹配
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // 移除通配符，保留点
		if strings.HasSuffix(domain, suffix) {
			// 确保通配符只匹配一个子域名
			prefix := strings.TrimSuffix(domain, suffix)
			prefix = strings.TrimSuffix(prefix, ".")
			if !strings.Contains(prefix, ".") {
				return true
			}
		}
	}

	return false
}

// parsePEMCertificate 解析 PEM 格式证书为 x509.Certificate
func parsePEMCertificate(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("invalid PEM block type: %s", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// extractCertInfo 从 x509.Certificate 提取证书信息
func extractCertInfo(cert *x509.Certificate, domain string) *CertInfo {
	now := time.Now()
	daysUntilExpiry := int(cert.NotAfter.Sub(now).Hours() / 24)

	// 计算指纹
	fingerprint := sha256.Sum256(cert.Raw)
	fingerprintStr := hex.EncodeToString(fingerprint[:])

	// 获取密钥算法
	keyAlgorithm := getKeyAlgorithm(cert)

	return &CertInfo{
		Domain:          domain,
		Issuer:          cert.Issuer.String(),
		NotBefore:       cert.NotBefore,
		NotAfter:        cert.NotAfter,
		Fingerprint:     fingerprintStr,
		SerialNumber:    cert.SerialNumber.String(),
		SANs:            cert.DNSNames,
		KeyAlgorithm:    keyAlgorithm,
		IsExpired:       now.After(cert.NotAfter),
		DaysUntilExpiry: daysUntilExpiry,
	}
}

// getKeyAlgorithm 获取证书密钥算法
func getKeyAlgorithm(cert *x509.Certificate) string {
	switch cert.PublicKeyAlgorithm {
	case x509.RSA:
		return "RSA"
	case x509.ECDSA:
		return "ECDSA"
	case x509.Ed25519:
		return "Ed25519"
	default:
		return "Unknown"
	}
}
