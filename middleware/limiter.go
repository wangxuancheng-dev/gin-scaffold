package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gin-scaffold/api/response"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/pkg/limiter"
)

// Limiter 基于 IP 与路由的令牌桶限流。
func Limiter(ipRPS float64, ipBurst int, routeRPS float64, routeBurst int) gin.HandlerFunc {
	return LimiterWithBackend(limiter.NewStore(ipRPS, ipBurst, routeRPS, routeBurst))
}

// LimiterWithStore 使用内存令牌桶限流（等价于 LimiterWithBackend(NewStore(...))）。
func LimiterWithStore(store *limiter.Store) gin.HandlerFunc {
	return LimiterWithBackend(store)
}

// LimiterWithBackend 使用指定限流后端（内存或 Redis 等）。
func LimiterWithBackend(b limiter.Backend) gin.HandlerFunc {
	return func(c *gin.Context) {
		if b == nil {
			c.Next()
			return
		}
		if !b.AllowIP(c.ClientIP()) {
			response.FailHTTP(c, http.StatusTooManyRequests, errcode.TooManyReq, errcode.KeyRateLimited, "too many requests")
			c.Abort()
			return
		}
		key := c.Request.Method + " " + c.FullPath()
		if key == " " {
			key = c.Request.Method + " " + c.Request.URL.Path
		}
		if !b.AllowRoute(key) {
			response.FailHTTP(c, http.StatusTooManyRequests, errcode.TooManyReq, errcode.KeyRateLimited, "route rate limited")
			c.Abort()
			return
		}
		c.Next()
	}
}
