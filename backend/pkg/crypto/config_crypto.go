package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// EncryptedPrefix 加密值的前缀标识
	EncryptedPrefix = "ENC("
	EncryptedSuffix = ")"
	// PBKDF2 参数
	pbkdf2Iterations = 10000
	keyLength        = 32 // AES-256 需要 32 字节密钥
	saltLength       = 16 // PBKDF2 salt 长度
)

// EncryptConfig 使用 salt 加密明文配置值
// 加密流程：
// 1. 生成随机 salt 用于 PBKDF2
// 2. 使用 PBKDF2 从用户提供的 salt 派生 32 字节 AES 密钥
// 3. 使用 AES-256-GCM 加密
// 4. 密文 base64 编码后包裹为 ENC(base64密文) 格式
func EncryptConfig(plaintext, userSalt string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	if userSalt == "" {
		return "", errors.New("salt cannot be empty")
	}

	// 生成随机 salt 用于 PBKDF2
	pbkdf2Salt := make([]byte, saltLength)
	if _, err := rand.Read(pbkdf2Salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// 使用 PBKDF2 派生密钥
	key := pbkdf2.Key([]byte(userSalt), pbkdf2Salt, pbkdf2Iterations, keyLength, sha256.New)

	// AES-256-GCM 加密
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 生成 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 加密：pbkdf2Salt + nonce + ciphertext
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// 组合：pbkdf2Salt(16) + nonce(12) + ciphertext
	result := make([]byte, 0, saltLength+len(nonce)+len(ciphertext))
	result = append(result, pbkdf2Salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	// Base64 编码并包裹
	encoded := base64.StdEncoding.EncodeToString(result)
	return EncryptedPrefix + encoded + EncryptedSuffix, nil
}

// DecryptConfig 使用 salt 解密密文配置值
func DecryptConfig(ciphertext, userSalt string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	if userSalt == "" {
		return "", errors.New("salt cannot be empty")
	}

	// 检查并移除 ENC(...) 包裹
	if !IsEncrypted(ciphertext) {
		return "", errors.New("value is not in encrypted format ENC(...)")
	}

	inner := ciphertext[len(EncryptedPrefix) : len(ciphertext)-1]

	// Base64 解码
	data, err := base64.StdEncoding.DecodeString(inner)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// 解析：pbkdf2Salt(16) + nonce(12) + ciphertext
	if len(data) < saltLength {
		return "", errors.New("ciphertext too short")
	}

	pbkdf2Salt := data[:saltLength]
	data = data[saltLength:]

	// 使用 PBKDF2 派生密钥
	key := pbkdf2.Key([]byte(userSalt), pbkdf2Salt, pbkdf2Iterations, keyLength, sha256.New)

	// AES-256-GCM 解密
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short for nonce")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// IsEncrypted 检查值是否为加密格式 ENC(...)
func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, EncryptedPrefix) && strings.HasSuffix(value, EncryptedSuffix)
}
