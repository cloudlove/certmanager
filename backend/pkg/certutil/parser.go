package certutil

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// ExtractCertInfo 从 PEM 提取所有证书信息 (域名/过期时间/颁发者/指纹/序列号等)
func ExtractCertInfo(certPEM string) (*CertInfo, error) {
	return ParseCertificate(certPEM)
}

// ParseCertChain 解析证书链，返回所有证书信息
func ParseCertChain(chainPEM string) ([]*CertInfo, error) {
	var certs []*CertInfo

	rest := []byte(chainPEM)
	for {
		block, remaining := pem.Decode(rest)
		if block == nil {
			break
		}

		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate in chain: %w", err)
			}

			// 尝试从证书中提取域名
			domain := ""
			if len(cert.DNSNames) > 0 {
				domain = cert.DNSNames[0]
			} else if cert.Subject.CommonName != "" {
				domain = cert.Subject.CommonName
			}

			certs = append(certs, extractCertInfo(cert, domain))
		}

		rest = remaining
		if len(rest) == 0 {
			break
		}
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("no valid certificates found in chain")
	}

	return certs, nil
}

// GetCertExpiryDate 获取证书过期时间
func GetCertExpiryDate(certPEM string) (int64, error) {
	certInfo, err := ParseCertificate(certPEM)
	if err != nil {
		return 0, err
	}
	return certInfo.NotAfter.Unix(), nil
}

// GetCertIssuer 获取证书颁发者
func GetCertIssuer(certPEM string) (string, error) {
	cert, err := parsePEMCertificate(certPEM)
	if err != nil {
		return "", err
	}
	return cert.Issuer.String(), nil
}

// GetCertSubject 获取证书主题
func GetCertSubject(certPEM string) (string, error) {
	cert, err := parsePEMCertificate(certPEM)
	if err != nil {
		return "", err
	}
	return cert.Subject.String(), nil
}

// GetCertDomains 获取证书包含的所有域名
func GetCertDomains(certPEM string) ([]string, error) {
	cert, err := parsePEMCertificate(certPEM)
	if err != nil {
		return nil, err
	}

	// 合并 Common Name 和 SANs
	domainMap := make(map[string]bool)

	if cert.Subject.CommonName != "" {
		domainMap[cert.Subject.CommonName] = true
	}

	for _, san := range cert.DNSNames {
		domainMap[san] = true
	}

	// 转换为切片
	domains := make([]string, 0, len(domainMap))
	for domain := range domainMap {
		domains = append(domains, domain)
	}

	return domains, nil
}

// ValidateCertPEM 验证 PEM 格式证书是否有效
func ValidateCertPEM(certPEM string) error {
	_, err := parsePEMCertificate(certPEM)
	return err
}

// IsSelfSigned 检查证书是否为自签名证书
func IsSelfSigned(certPEM string) (bool, error) {
	cert, err := parsePEMCertificate(certPEM)
	if err != nil {
		return false, err
	}

	// 检查 Issuer 和 Subject 是否相同
	return cert.Issuer.String() == cert.Subject.String(), nil
}
