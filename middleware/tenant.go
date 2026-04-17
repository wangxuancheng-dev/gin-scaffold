package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"gin-scaffold/config"
	"gin-scaffold/internal/pkg/tenant"
)

func Tenant(cfg *config.TenantConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg == nil || !cfg.Enabled {
			c.Next()
			return
		}
		header := strings.TrimSpace(cfg.Header)
		if header == "" {
			header = "X-Tenant-ID"
		}
		tenantID := strings.TrimSpace(c.GetHeader(header))
		if tenantID == "" {
			if claims, ok := Claims(c); ok && claims != nil {
				tenantID = strings.TrimSpace(claims.TenantID)
			}
		}
		if tenantID == "" {
			tenantID = strings.TrimSpace(cfg.DefaultID)
		}
		if tenantID != "" {
			c.Request = c.Request.WithContext(tenant.WithContext(c.Request.Context(), tenantID))
			c.Set("tenant_id", tenantID)
			c.Writer.Header().Set("X-Tenant-ID", tenantID)
		}
		c.Next()
	}
}
