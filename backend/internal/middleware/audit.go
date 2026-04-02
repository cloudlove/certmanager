package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"certmanager-backend/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuditMiddleware 审计日志中间件
func AuditMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只记录写操作
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete {
			c.Next()
			return
		}

		// 获取请求信息
		path := c.Request.URL.Path
		action := getActionFromMethod(method)
		resourceType := getResourceTypeFromPath(path)

		// 读取请求体
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 获取资源ID（从路径或请求体）
		resourceID := getResourceIDFromPath(path)
		if resourceID == 0 && len(bodyBytes) > 0 {
			resourceID = getResourceIDFromBody(bodyBytes)
		}

		// 获取客户端IP
		clientIP := c.ClientIP()

		// 获取用户ID（从上下文或token，暂时使用IP作为标识）
		userID := c.GetString("userId")
		if userID == "" {
			userID = clientIP
		}

		// 构建详情
		detail := buildDetail(path, method, bodyBytes)

		// 异步记录审计日志
		go func() {
			auditLog := &model.AuditLog{
				UserID:       userID,
				Action:       action,
				ResourceType: resourceType,
				ResourceID:   resourceID,
				Detail:       detail,
				IP:           clientIP,
			}
			db.Create(auditLog)
		}()

		c.Next()
	}
}

// getActionFromMethod 根据HTTP方法获取操作类型
func getActionFromMethod(method string) string {
	switch method {
	case http.MethodPost:
		return "create"
	case http.MethodPut:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return "unknown"
	}
}

// getResourceTypeFromPath 从路径获取资源类型
func getResourceTypeFromPath(path string) string {
	path = strings.TrimPrefix(path, "/api/v1/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return "unknown"
	}

	switch parts[0] {
	case "certificates":
		return "certificate"
	case "credentials":
		return "credential"
	case "csrs":
		return "csr"
	case "domains":
		return "domain"
	case "deployments":
		return "deploy_task"
	case "nginx":
		if len(parts) > 1 && parts[1] == "clusters" {
			return "nginx_cluster"
		}
		return "nginx"
	case "notifications":
		return "notification_rule"
	default:
		return parts[0]
	}
}

// getResourceIDFromPath 从路径获取资源ID
func getResourceIDFromPath(path string) uint {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if i > 0 && (parts[i-1] == "certificates" ||
			parts[i-1] == "credentials" ||
			parts[i-1] == "csrs" ||
			parts[i-1] == "domains" ||
			parts[i-1] == "deployments" ||
			parts[i-1] == "clusters" ||
			parts[i-1] == "nodes" ||
			parts[i-1] == "notifications") {
			// 尝试解析ID
			var id uint
			if _, err := json.Number(part).Int64(); err == nil {
				// 简单解析数字
				for _, ch := range part {
					if ch >= '0' && ch <= '9' {
						id = id*10 + uint(ch-'0')
					} else {
						break
					}
				}
				return id
			}
		}
	}
	return 0
}

// getResourceIDFromBody 从请求体获取资源ID
func getResourceIDFromBody(body []byte) uint {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return 0
	}

	if id, ok := data["id"].(float64); ok {
		return uint(id)
	}
	return 0
}

// buildDetail 构建操作详情
func buildDetail(path, method string, body []byte) string {
	var sb strings.Builder
	sb.WriteString("Path: ")
	sb.WriteString(path)
	sb.WriteString(", Method: ")
	sb.WriteString(method)

	if len(body) > 0 {
		// 脱敏处理后的请求体
		bodyStr := sanitizeSensitiveData(string(body))
		if len(bodyStr) > 500 {
			bodyStr = bodyStr[:500] + "..."
		}
		sb.WriteString(", Body: ")
		sb.WriteString(bodyStr)
	}

	return sb.String()
}

// sanitizeSensitiveData 对敏感数据进行脱敏处理
func sanitizeSensitiveData(body string) string {
	// 定义需要脱敏的敏感字段
	sensitiveFields := []string{
		"password", "old_password", "new_password",
		"secret", "token", "access_token", "refresh_token",
		"api_key", "apikey", "key", "private_key",
		"credential", "passwd", "pwd",
	}

	result := body
	for _, field := range sensitiveFields {
		// 匹配 "field": "value" 或 'field': 'value' 格式
		patterns := []string{
			`"` + field + `"\s*:\s*"[^"]*"`,
			`"` + field + `"\s*:\s*'[^']*'`,
			`'` + field + `'\s*:\s*"[^"]*"`,
			`'` + field + `'\s*:\s*'[^']*'`,
		}
		for _, pattern := range patterns {
			result = sanitizeField(result, pattern)
		}
	}
	return result
}

// sanitizeField 脱敏单个字段
func sanitizeField(input, pattern string) string {
	// 简单替换：将值替换为 "***"
	parts := strings.Split(input, pattern)
	if len(parts) <= 1 {
		return input
	}

	// 找到匹配的字段并替换值
	fieldEnd := strings.Index(pattern, `:`)
	if fieldEnd == -1 {
		return input
	}
	fieldName := pattern[:fieldEnd]
	replacement := fieldName + `: "***"`

	return strings.ReplaceAll(input, pattern, replacement)
}
