package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery panic 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 打印堆栈信息
				fmt.Printf("[PANIC] %v\n%s\n", err, string(debug.Stack()))

				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
