package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"gin-scaffold/api/response"
	"gin-scaffold/internal/pkg/errcode"
)

// BodyLimit 限制请求体大小，单位字节；<=0 时使用默认 1MB。
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	if maxBytes <= 0 {
		maxBytes = 1 << 20
	}
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			response.FailHTTP(
				c,
				http.StatusRequestEntityTooLarge,
				errcode.PayloadTooLarge,
				errcode.KeyPayloadTooLarge,
				"payload too large",
			)
			c.Abort()
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
		for _, e := range c.Errors {
			var maxErr *http.MaxBytesError
			if errors.As(e.Err, &maxErr) {
				if !c.Writer.Written() {
					response.FailHTTP(
						c,
						http.StatusRequestEntityTooLarge,
						errcode.PayloadTooLarge,
						errcode.KeyPayloadTooLarge,
						"payload too large",
					)
				}
				c.Abort()
				return
			}
		}
	}
}
