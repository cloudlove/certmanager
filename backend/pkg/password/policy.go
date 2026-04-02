package password

import (
	"errors"
	"unicode"
)

// Policy 密码策略配置
type Policy struct {
	MinLength          int  // 最小长度
	RequireUppercase   bool // 需要大写字母
	RequireLowercase   bool // 需要小写字母
	RequireDigit       bool // 需要数字
	RequireSpecialChar bool // 需要特殊字符
}

// DefaultPolicy 默认密码策略
var DefaultPolicy = Policy{
	MinLength:          8,
	RequireUppercase:   true,
	RequireLowercase:   true,
	RequireDigit:       true,
	RequireSpecialChar: false,
}

// ValidatePasswordStrength 验证密码强度
// 默认策略：最少8位，需包含大小写字母和数字
func ValidatePasswordStrength(password string) error {
	return ValidatePasswordStrengthWithPolicy(password, DefaultPolicy)
}

// ValidatePasswordStrengthWithPolicy 使用指定策略验证密码强度
func ValidatePasswordStrengthWithPolicy(password string, policy Policy) error {
	if len(password) < policy.MinLength {
		return errors.New("密码长度至少为8位")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if policy.RequireUppercase && !hasUpper {
		return errors.New("密码必须包含大写字母")
	}

	if policy.RequireLowercase && !hasLower {
		return errors.New("密码必须包含小写字母")
	}

	if policy.RequireDigit && !hasDigit {
		return errors.New("密码必须包含数字")
	}

	if policy.RequireSpecialChar && !hasSpecial {
		return errors.New("密码必须包含特殊字符")
	}

	return nil
}

// IsStrongPassword 检查密码是否为强密码（返回布尔值）
func IsStrongPassword(password string) bool {
	return ValidatePasswordStrength(password) == nil
}
