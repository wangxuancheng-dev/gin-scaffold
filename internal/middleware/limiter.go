package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"gin-scaffold/internal/api/response"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/pkg/limiter"
)

// LimiterKeys 自定义限流维度键
// 字段为 nil 时使用默认：IP 为 c.ClientIP()，路由为「方法 + 空格 + FullPath（或 Path）」。
type LimiterKeys struct {
	IPKey    func(*gin.Context) string
	RouteKey func(*gin.Context) string
}

// Limiter 基于 IP 与路由的令牌桶限流。
func Limiter(ipRPS float64, ipBurst int, routeRPS float64, routeBurst int) gin.HandlerFunc {
	return LimiterWithBackend(limiter.NewStore(ipRPS, ipBurst, routeRPS, routeBurst))
}

// LimiterWithStore 使用内存令牌桶限流（等价于 LimiterWithBackend(NewStore(...))）。
func LimiterWithStore(store *limiter.Store) gin.HandlerFunc {
	return LimiterWithBackend(store)
}

// LimiterWithStoreKeys 使用内存后端并自定义 IP / 路由键。
func LimiterWithStoreKeys(store *limiter.Store, keys *LimiterKeys) gin.HandlerFunc {
	return LimiterWithBackendKeys(store, keys)
}

// LimiterWithBackend 使用指定限流后端（内存或 Redis 等）。
func LimiterWithBackend(b limiter.Backend) gin.HandlerFunc {
	return LimiterWithBackendKeys(b, nil)
}

// LimiterWithBackendKeys 与 LimiterWithBackend 相同，但可通过 keys 自定义 AllowIP / AllowRoute 使用的字符串键。
func LimiterWithBackendKeys(b limiter.Backend, keys *LimiterKeys) gin.HandlerFunc {
	return func(c *gin.Context) {
		if b == nil {
			c.Next()
			return
		}
		ipKey := c.ClientIP()
		if keys != nil && keys.IPKey != nil {
			if k := strings.TrimSpace(keys.IPKey(c)); k != "" {
				ipKey = k
			}
		}
		if !b.AllowIP(ipKey) {
			response.FailHTTP(c, http.StatusTooManyRequests, errcode.TooManyReq, errcode.KeyRateLimited, "too many requests")
			c.Abort()
			return
		}
		routeKey := defaultLimiterRouteKey(c)
		if keys != nil && keys.RouteKey != nil {
			if k := strings.TrimSpace(keys.RouteKey(c)); k != "" {
				routeKey = k
			}
		}
		if !b.AllowRoute(routeKey) {
			response.FailHTTP(c, http.StatusTooManyRequests, errcode.TooManyReq, errcode.KeyRateLimited, "route rate limited")
			c.Abort()
			return
		}
		c.Next()
	}
}

func defaultLimiterRouteKey(c *gin.Context) string {
	key := c.Request.Method + " " + c.FullPath()
	if key == " " {
		key = c.Request.Method + " " + c.Request.URL.Path
	}
	return key
}
