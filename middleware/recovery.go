package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-scaffold/api/response"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/pkg/logger"
)

// Recovery 捕获 Panic，返回 500 并记录堆栈。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.ErrorX("panic recovered",
					zap.Any("panic", rec),
					zap.String("request_id", GetRequestID(c)),
					zap.ByteString("stack", debug.Stack()),
				)
				response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, "internal error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
