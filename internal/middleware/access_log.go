package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-scaffold/pkg/logger"
)

// AccessLog 记录 HTTP 访问日志（方法、路径、耗时、状态码、IP）。
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		cost := time.Since(start)
		if logger.Access() != nil {
			logger.Access().Info("access",
				zap.String("request_id", GetRequestID(c)),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("latency", cost),
				zap.String("client_ip", c.ClientIP()),
			)
		}
	}
}
