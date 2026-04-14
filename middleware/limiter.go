package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"gin-scaffold/api/response"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/pkg/limiter"
)

var globalLimiter *limiter.Store
var limiterOnce sync.Once

// Limiter 基于 IP 与路由的令牌桶限流。
func Limiter(ipRPS float64, ipBurst int, routeRPS float64, routeBurst int) gin.HandlerFunc {
	limiterOnce.Do(func() {
		globalLimiter = limiter.NewStore(ipRPS, ipBurst, routeRPS, routeBurst)
	})
	return func(c *gin.Context) {
		if globalLimiter == nil {
			c.Next()
			return
		}
		if !globalLimiter.AllowIP(c.ClientIP()) {
			response.FailHTTP(c, http.StatusTooManyRequests, errcode.BadRequest, errcode.KeyInvalidParam, "too many requests")
			c.Abort()
			return
		}
		key := c.Request.Method + " " + c.FullPath()
		if key == " " {
			key = c.Request.Method + " " + c.Request.URL.Path
		}
		if !globalLimiter.AllowRoute(key) {
			response.FailHTTP(c, http.StatusTooManyRequests, errcode.BadRequest, errcode.KeyInvalidParam, "route rate limited")
			c.Abort()
			return
		}
		c.Next()
	}
}
