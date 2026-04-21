package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"gin-scaffold/internal/config"
)

// CORS 根据配置启用跨域。
func CORS(cfg *config.CORSConfig) gin.HandlerFunc {
	if cfg == nil {
		return func(c *gin.Context) { c.Next() }
	}
	allowMethods := cfg.AllowMethods
	if len(allowMethods) == 0 {
		allowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	allowHeaders := cfg.AllowHeaders
	if len(allowHeaders) == 0 {
		allowHeaders = []string{"Authorization", "Content-Type", "Accept-Language", "X-Request-ID"}
	}
	exposeHeaders := cfg.ExposeHeaders
	if len(exposeHeaders) == 0 {
		exposeHeaders = []string{"Content-Length", "X-Request-ID"}
	}
	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		ExposeHeaders:    exposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           12 * time.Hour,
	})
}
