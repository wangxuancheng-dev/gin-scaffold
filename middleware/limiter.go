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
	return LimiterWithStore(limiter.NewStore(ipRPS, ipBurst, routeRPS, routeBurst))
}

// LimiterWithStore 使用指定限流实例构建中间件（便于测试与隔离）。
func LimiterWithStore(store *limiter.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if store == nil {
			c.Next()
			return
		}
		if !store.AllowIP(c.ClientIP()) {
			response.FailHTTP(c, http.StatusTooManyRequests, errcode.TooManyReq, errcode.KeyRateLimited, "too many requests")
			c.Abort()
			return
		}
		key := c.Request.Method + " " + c.FullPath()
		if key == " " {
			key = c.Request.Method + " " + c.Request.URL.Path
		}
		if !store.AllowRoute(key) {
			response.FailHTTP(c, http.StatusTooManyRequests, errcode.TooManyReq, errcode.KeyRateLimited, "route rate limited")
			c.Abort()
			return
		}
		c.Next()
	}
}
