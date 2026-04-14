// Package middleware 提供 Gin 通用中间件（跨域、鉴权、日志、限流等）。
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const headerRequestID = "X-Request-ID"
const ctxRequestID = "request_id"

// RequestID 注入或透传请求唯一 ID。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(headerRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Writer.Header().Set(headerRequestID, rid)
		c.Set(ctxRequestID, rid)
		c.Next()
	}
}

// GetRequestID 从上下文读取 RequestID。
func GetRequestID(c *gin.Context) string {
	v, ok := c.Get(ctxRequestID)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}
