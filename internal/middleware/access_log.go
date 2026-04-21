package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"gin-scaffold/pkg/logger"
)

// AccessLog 记录 HTTP 访问日志（方法、路径、耗时、状态码、IP）。
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		cost := time.Since(start)
		log := logger.Access()
		if log == nil {
			return
		}
		traceID := ""
		if sc := trace.SpanContextFromContext(c.Request.Context()); sc.IsValid() {
			traceID = sc.TraceID().String()
		}
		userID := int64(0)
		role := ""
		if claims, ok := Claims(c); ok && claims != nil {
			userID = claims.UserID
			role = claims.Role
		}
		tenantID := ""
		if v, ok := c.Get("tenant_id"); ok {
			if s, cast := v.(string); cast {
				tenantID = s
			}
		}
		fields := []zap.Field{
			zap.String("request_id", GetRequestID(c)),
			zap.String("trace_id", traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", cost),
			zap.String("client_ip", c.ClientIP()),
			zap.Int64("user_id", userID),
			zap.String("role", role),
			zap.String("tenant_id", tenantID),
		}
		if cost >= time.Second {
			log.Warn("slow_access", fields...)
			return
		}
		log.Info("access", fields...)
	}
}
