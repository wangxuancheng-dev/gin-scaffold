package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gin-scaffold/internal/api/response"
	"gin-scaffold/internal/config"
	appredis "gin-scaffold/pkg/redis"
	"gin-scaffold/pkg/storage"
)

// BaseHandler 健康检查等基础接口。
type BaseHandler struct {
	DB      *gorm.DB
	Storage *config.StorageConfig
}

// Health 兼容入口，等价于 Readyz（依赖就绪检查）。
// @Summary 健康检查（兼容）
// @Description 等价于 /readyz
// @Tags base
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *BaseHandler) Health(c *gin.Context) {
	h.Readyz(c)
}

// Livez 进程存活检查，仅表示服务进程可响应。
// @Summary 存活检查
// @Description 仅检测应用进程自身是否存活
// @Tags base
// @Produce json
// @Success 200 {object} response.Body
// @Router /livez [get]
func (h *BaseHandler) Livez(c *gin.Context) {
	response.OK(c, gin.H{"app": "ok"})
}

// Readyz 依赖就绪检查：数据库、Redis；若配置 storage.readyz_check=true 则额外检查存储连通性。
// @Summary 就绪检查
// @Description 返回依赖组件就绪状态
// @Tags base
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /readyz [get]
func (h *BaseHandler) Readyz(c *gin.Context) {
	out, ok := h.readiness(c)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, out)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *BaseHandler) readiness(c *gin.Context) (gin.H, bool) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	out := gin.H{"app": "ok", "db": "skip", "redis": "skip"}
	if h.DB != nil {
		sqlDB, err := h.DB.DB()
		if err != nil {
			out["db"] = err.Error()
		} else if err := sqlDB.PingContext(ctx); err != nil {
			out["db"] = err.Error()
		} else {
			out["db"] = "ok"
		}
	}
	if err := appredis.Ping(ctx); err != nil {
		out["redis"] = err.Error()
		return out, false
	}
	out["redis"] = "ok"
	if out["db"] != "ok" && out["db"] != "skip" {
		return out, false
	}
	out["storage"] = "skip"
	if h.Storage != nil && h.Storage.ReadyzCheck {
		p, err := storage.Require()
		if err != nil {
			out["storage"] = err.Error()
			return out, false
		}
		rc, ok := p.(storage.ReadinessChecker)
		if !ok {
			out["storage"] = "readiness unsupported for storage driver"
			return out, false
		}
		if err := rc.Ready(ctx); err != nil {
			out["storage"] = err.Error()
			return out, false
		}
		out["storage"] = "ok"
	}
	return out, true
}

// Ping 简单存活探测。
// @Summary Ping
// @Tags base
// @Success 200 {object} response.Body
// @Router /api/v1/client/ping [get]
func (h *BaseHandler) Ping(c *gin.Context) {
	response.OK(c, gin.H{"pong": true})
}
