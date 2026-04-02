package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType 令牌类型
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// Claims JWT 声明
type Claims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	RoleID    uint   `json:"role_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// GenerateTokenPair 生成访问令牌和刷新令牌
func GenerateTokenPair(userID uint, username string, roleID uint, secret string, expireHours, refreshHours int) (string, string, error) {
	accessToken, err := GenerateToken(userID, username, roleID, secret, expireHours, TokenTypeAccess)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := GenerateToken(userID, username, roleID, secret, refreshHours, TokenTypeRefresh)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// GenerateToken 生成 JWT 令牌
func GenerateToken(userID uint, username string, roleID uint, secret string, expireHours int, tokenType string) (string, error) {
	claims := Claims{
		UserID:    userID,
		Username:  username,
		RoleID:    roleID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "certmanager",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析 JWT 令牌
func ParseToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// GetTokenExpireTime 获取令牌过期时间
func GetTokenExpireTime(tokenString string, secret string) (time.Time, error) {
	claims, err := ParseToken(tokenString, secret)
	if err != nil {
		return time.Time{}, err
	}
	return claims.ExpiresAt.Time, nil
}
