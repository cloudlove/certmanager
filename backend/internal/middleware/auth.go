package middleware

import (
	"strings"

	"certmanager-backend/internal/config"
	"certmanager-backend/pkg/jwt"
	"certmanager-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// 上下文键名
const (
	ContextKeyUserID   = "user_id"
	ContextKeyUsername = "username"
	ContextKeyRoleID   = "role_id"
)

// AuthMiddleware JWT 令牌验证中间件
// 从 Authorization: Bearer xxx 提取并验证 token，将用户信息存入 gin.Context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取 Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, 401, "missing authorization header")
			c.Abort()
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, 401, "invalid authorization format, expected 'Bearer <token>'")
			c.Abort()
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			response.Error(c, 401, "token is empty")
			c.Abort()
			return
		}

		// 获取 JWT 配置
		cfg := config.GetConfig()
		secret := cfg.JWT.Secret

		// 解析 token
		claims, err := jwt.ParseToken(tokenString, secret)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				response.Error(c, 401, "token expired")
			} else {
				response.Error(c, 401, "invalid token")
			}
			c.Abort()
			return
		}

		// 检查 token 类型，只允许 access token
		if claims.TokenType != jwt.TokenTypeAccess {
			response.Error(c, 401, "invalid token type, expected access token")
			c.Abort()
			return
		}

		// 将用户信息存入 context
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyRoleID, claims.RoleID)

		c.Next()
	}
}

// PermissionMiddleware 权限检查中间件
// 从 context 获取用户角色，检查是否有对应权限
// 需要 PermissionService 注入
type PermissionChecker interface {
	HasPermission(roleID uint, resource, action string) bool
}

var permissionChecker PermissionChecker

// SetPermissionChecker 设置权限检查器
func SetPermissionChecker(checker PermissionChecker) {
	permissionChecker = checker
}

// PermissionMiddleware 权限检查中间件
func PermissionMiddleware(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色 ID
		roleIDVal, exists := c.Get(ContextKeyRoleID)
		if !exists {
			response.Error(c, 401, "unauthorized: user not found in context")
			c.Abort()
			return
		}

		roleID, ok := roleIDVal.(uint)
		if !ok {
			response.Error(c, 500, "invalid role id type")
			c.Abort()
			return
		}

		// 检查权限
		if permissionChecker == nil {
			response.Error(c, 500, "permission checker not initialized")
			c.Abort()
			return
		}

		if !permissionChecker.HasPermission(roleID, resource, action) {
			response.Error(c, 403, "forbidden: no permission to "+action+" "+resource)
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUser 从 context 获取当前用户信息
func GetCurrentUser(c *gin.Context) (userID uint, username string, roleID uint, ok bool) {
	userIDVal, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, "", 0, false
	}
	usernameVal, exists := c.Get(ContextKeyUsername)
	if !exists {
		return 0, "", 0, false
	}
	roleIDVal, exists := c.Get(ContextKeyRoleID)
	if !exists {
		return 0, "", 0, false
	}

	userID, ok1 := userIDVal.(uint)
	username, ok2 := usernameVal.(string)
	roleID, ok3 := roleIDVal.(uint)

	if !ok1 || !ok2 || !ok3 {
		return 0, "", 0, false
	}

	return userID, username, roleID, true
}
