package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s | %3d | %13v | %15s | %s %s\n",
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Path,
			param.ErrorMessage,
		)
	})
}
