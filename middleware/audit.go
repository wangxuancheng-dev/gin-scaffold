package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-scaffold/config"
	"gin-scaffold/internal/dao"
	"gin-scaffold/internal/model"
	"gin-scaffold/pkg/db"
	"gin-scaffold/pkg/logger"
)

// Audit 将变更类 HTTP 请求异步写入 audit_logs（需执行迁移且 platform.audit.enabled=true）。
func Audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get()
		if cfg == nil || !cfg.Platform.Audit.Enabled {
			c.Next()
			return
		}
		start := time.Now()
		c.Next()

		switch c.Request.Method {
		case "POST", "PUT", "PATCH", "DELETE":
		default:
			return
		}
		path := c.Request.URL.Path
		if auditSkipPath(path) {
			return
		}

		gdb := db.DB()
		if gdb == nil {
			return
		}

		row := buildAuditRow(c, c.Request.Method, path, start)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := dao.NewAuditLogDAO(gdb).Create(ctx, row); err != nil {
				logger.L().Warn("audit insert failed", zap.Error(err), zap.String("path", path))
			}
		}()
	}
}

func auditSkipPath(path string) bool {
	if path == "/livez" || path == "/readyz" || path == "/health" {
		return true
	}
	if strings.HasPrefix(path, "/swagger") || strings.HasPrefix(path, "/debug") {
		return true
	}
	if path == "/metrics" || strings.HasPrefix(path, "/metrics/") {
		return true
	}
	return false
}

func buildAuditRow(c *gin.Context, method, path string, start time.Time) *model.AuditLog {
	latency := int(time.Since(start).Milliseconds())
	if latency < 0 {
		latency = 0
	}
	status := c.Writer.Status()
	if status == 0 {
		status = 200
	}
	row := &model.AuditLog{
		RequestID: GetRequestID(c),
		Action:    method,
		Path:      path,
		Query:     truncateStr(c.Request.URL.RawQuery, 1024),
		Status:    status,
		LatencyMS: latency,
		ClientIP:  truncateStr(c.ClientIP(), 64),
		CreatedAt: time.Now(),
	}
	if cl, ok := Claims(c); ok && cl != nil {
		row.UserID = cl.UserID
		row.Role = truncateStr(cl.Role, 32)
		row.ActorType = "jwt"
	} else {
		row.ActorType = "anonymous"
	}
	return row
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
