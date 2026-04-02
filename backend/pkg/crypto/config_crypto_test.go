package crypto

import (
	"fmt"
	"testing"
)

func TestEncryptDecryptConfig(t *testing.T) {
	salt := "certmanager-default-salt"
	testCases := []string{
		"changeme",
		"changeme-aes-key-must-be-32bytes",
		"changeme-jwt-secret-must-be-32b",
		"", // 空值
	}

	for _, tc := range testCases {
		encrypted, err := EncryptConfig(tc, salt)
		if err != nil {
			t.Errorf("EncryptConfig(%q) error: %v", tc, err)
			continue
		}

		if tc == "" {
			if encrypted != "" {
				t.Errorf("EncryptConfig('') should return '', got %q", encrypted)
			}
			continue
		}

		fmt.Printf("明文: %q\n加密结果: %s\n\n", tc, encrypted)

		decrypted, err := DecryptConfig(encrypted, salt)
		if err != nil {
			t.Errorf("DecryptConfig error: %v", err)
			continue
		}

		if decrypted != tc {
			t.Errorf("DecryptConfig mismatch: got %q, want %q", decrypted, tc)
		}
	}
}

func TestIsEncrypted(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"ENC(abc123)", true},
		{"ENC(base64encoded==)", true},
		{"plain text", false},
		{"ENC(incomplete", false},
		{"incomplete)", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsEncrypted(tt.value)
		if result != tt.expected {
			t.Errorf("IsEncrypted(%q) = %v, want %v", tt.value, result, tt.expected)
		}
	}
}
