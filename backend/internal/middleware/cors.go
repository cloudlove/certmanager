package middleware

import (
	"net/http"
	"strings"

	"certmanager-backend/internal/config"

	"github.com/gin-gonic/gin"
)

// CORS CORS 中间件（支持白名单模式）
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		cfg := config.GetConfig()

		// 检查 origin 是否在白名单中
		allowedOrigin := ""
		if len(cfg.CORS.AllowedOrigins) == 0 {
			// 默认允许本地开发环境
			allowedOrigin = origin
		} else {
			for _, allowed := range cfg.CORS.AllowedOrigins {
				if allowed == "*" || strings.EqualFold(allowed, origin) {
					allowedOrigin = origin
					break
				}
			}
		}

		// 设置 CORS 响应头
		if allowedOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}

		// 只有在配置了允许 credentials 且不是通配符时才设置
		if cfg.CORS.AllowCredentials && allowedOrigin != "" && allowedOrigin != "*" {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
