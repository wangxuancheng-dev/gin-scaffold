package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-scaffold/api/response"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/pkg/logger"
)

// Recovery 捕获 Panic，返回 500 并记录堆栈。
// debug=true 时记录更多请求上下文信息，便于本地排障。
func Recovery(debugMode bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				fields := []zap.Field{
					zap.Any("panic", rec),
					zap.String("request_id", GetRequestID(c)),
					zap.ByteString("stack", debug.Stack()),
				}
				if debugMode {
					fields = append(fields,
						zap.String("method", c.Request.Method),
						zap.String("path", c.Request.URL.Path),
						zap.String("query", c.Request.URL.RawQuery),
						zap.String("client_ip", c.ClientIP()),
						zap.String("user_agent", c.Request.UserAgent()),
						zap.String("recover_hint", fmt.Sprintf("panic happened in debug mode, check request_id=%s", GetRequestID(c))),
					)
				}
				logger.ErrorX("panic recovered", fields...)
				response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, "internal error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
