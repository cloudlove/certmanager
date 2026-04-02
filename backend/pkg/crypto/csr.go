package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"net"
	"strings"
)

// CSRInfo CSR 信息结构体
type CSRInfo struct {
	CommonName   string   `json:"common_name"`
	SANs         []string `json:"sans"`
	KeyAlgorithm string   `json:"key_algorithm"`
	KeySize      string   `json:"key_size"`
	PublicKey    string   `json:"public_key"`
}

// GenerateCSR 生成 CSR 和私钥
// 支持: RSA 2048/4096, ECC P256/P384
func GenerateCSR(commonName string, sans []string, keyAlgorithm string, keySize string, countryCode, province, locality, corpName, department string) (csrPEM string, privateKeyPEM string, err error) {
	// 生成密钥对
	var privateKey crypto.PrivateKey

	switch strings.ToUpper(keyAlgorithm) {
	case "RSA":
		privateKey, _, err = generateRSAKey(keySize)
	case "ECC":
		privateKey, _, err = generateECCKey(keySize)
	default:
		return "", "", fmt.Errorf("unsupported key algorithm: %s", keyAlgorithm)
	}

	if err != nil {
		return "", "", fmt.Errorf("failed to generate key: %w", err)
	}

	// 创建 CSR 模板
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: commonName,
		},
	}

	// 设置组织信息
	if countryCode != "" {
		template.Subject.Country = []string{countryCode}
	}
	if province != "" {
		template.Subject.Province = []string{province}
	}
	if locality != "" {
		template.Subject.Locality = []string{locality}
	}
	if corpName != "" {
		template.Subject.Organization = []string{corpName}
	}
	if department != "" {
		template.Subject.OrganizationalUnit = []string{department}
	}

	// 添加 SAN 扩展
	if len(sans) > 0 {
		dnsNames, ipAddresses := parseSANs(sans)
		template.DNSNames = dnsNames
		template.IPAddresses = ipAddresses
	}

	// 生成 CSR
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create CSR: %w", err)
	}

	// 编码 CSR 为 PEM
	csrPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	}))

	// 编码私钥为 PEM
	privateKeyPEM, err = encodePrivateKeyToPEM(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode private key: %w", err)
	}

	return csrPEM, privateKeyPEM, nil
}

// ParseCSR 解析 CSR PEM 返回详细信息
func ParseCSR(csrPEM string) (*CSRInfo, error) {
	block, _ := pem.Decode([]byte(csrPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSR: %w", err)
	}

	// 验证 CSR 签名
	if err := csr.CheckSignature(); err != nil {
		return nil, fmt.Errorf("CSR signature verification failed: %w", err)
	}

	// 获取密钥算法信息
	keyAlgorithm, keySize := getKeyInfo(csr.PublicKey)

	// 提取 SANs
	sans := extractSANs(csr)

	return &CSRInfo{
		CommonName:   csr.Subject.CommonName,
		SANs:         sans,
		KeyAlgorithm: keyAlgorithm,
		KeySize:      keySize,
		PublicKey:    encodePublicKeyToPEM(csr.PublicKey),
	}, nil
}

// generateRSAKey 生成 RSA 密钥对
func generateRSAKey(keySize string) (crypto.PrivateKey, crypto.PublicKey, error) {
	var bits int
	switch keySize {
	case "2048":
		bits = 2048
	case "4096":
		bits = 4096
	default:
		bits = 2048
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, &privateKey.PublicKey, nil
}

// generateECCKey 生成 ECC 密钥对
func generateECCKey(keySize string) (crypto.PrivateKey, crypto.PublicKey, error) {
	var curve elliptic.Curve
	switch keySize {
	case "P256", "256":
		curve = elliptic.P256()
	case "P384", "384":
		curve = elliptic.P384()
	default:
		curve = elliptic.P256()
	}

	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, &privateKey.PublicKey, nil
}

// parseSANs 解析 SAN 列表，分离 DNS 名称和 IP 地址
func parseSANs(sans []string) (dnsNames []string, ipAddresses []net.IP) {
	for _, san := range sans {
		san = strings.TrimSpace(san)
		if san == "" {
			continue
		}
		// 检查是否为 IP 地址
		if ip := net.ParseIP(san); ip != nil {
			ipAddresses = append(ipAddresses, ip)
		} else {
			dnsNames = append(dnsNames, san)
		}
	}
	return
}

// encodePrivateKeyToPEM 将私钥编码为 PEM 格式
func encodePrivateKeyToPEM(privateKey crypto.PrivateKey) (string, error) {
	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		return string(pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		})), nil
	case *ecdsa.PrivateKey:
		bytes, err := x509.MarshalECPrivateKey(key)
		if err != nil {
			return "", err
		}
		return string(pem.EncodeToMemory(&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: bytes,
		})), nil
	default:
		return "", fmt.Errorf("unsupported private key type")
	}
}

// encodePublicKeyToPEM 将公钥编码为 PEM 格式
func encodePublicKeyToPEM(publicKey crypto.PublicKey) string {
	bytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return ""
	}
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: bytes,
	}))
}

// getKeyInfo 获取密钥算法和大小信息
func getKeyInfo(publicKey crypto.PublicKey) (algorithm string, keySize string) {
	switch key := publicKey.(type) {
	case *rsa.PublicKey:
		return "RSA", fmt.Sprintf("%d", key.N.BitLen())
	case *ecdsa.PublicKey:
		switch key.Curve {
		case elliptic.P256():
			return "ECC", "P256"
		case elliptic.P384():
			return "ECC", "P384"
		case elliptic.P521():
			return "ECC", "P521"
		default:
			return "ECC", "Unknown"
		}
	default:
		return "Unknown", "Unknown"
	}
}

// extractSANs 从 CSR 中提取 SAN 列表
func extractSANs(csr *x509.CertificateRequest) []string {
	var sans []string

	// 添加 DNS 名称
	sans = append(sans, csr.DNSNames...)

	// 添加 IP 地址
	for _, ip := range csr.IPAddresses {
		sans = append(sans, ip.String())
	}

	// 从扩展中提取 SAN
	for _, ext := range csr.Extensions {
		if ext.Id.Equal(asn1.ObjectIdentifier{2, 5, 29, 17}) { // subjectAltName OID
			var rawValues []asn1.RawValue
			if _, err := asn1.Unmarshal(ext.Value, &rawValues); err == nil {
				for _, rv := range rawValues {
					switch rv.Tag {
					case 2: // dNSName
						sans = append(sans, string(rv.Bytes))
					case 7: // iPAddress
						if len(rv.Bytes) == 4 || len(rv.Bytes) == 16 {
							sans = append(sans, net.IP(rv.Bytes).String())
						}
					}
				}
			}
		}
	}

	return sans
}
