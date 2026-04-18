package middleware

import (
	"github.com/gin-gonic/gin"

	"gin-scaffold/internal/pkg/clientip"
)

// ClientIPContext 将 Gin 解析到的客户端 IP 写入 request context。
func ClientIPContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(clientip.With(c.Request.Context(), c.ClientIP()))
		c.Next()
	}
}
