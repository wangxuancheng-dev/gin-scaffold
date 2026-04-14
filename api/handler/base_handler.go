package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gin-scaffold/api/response"
	appredis "gin-scaffold/pkg/redis"
)

// BaseHandler 健康检查等基础接口。
type BaseHandler struct {
	DB *gorm.DB
}

// Health 聚合健康检查：本进程、数据库、Redis。
// @Summary 健康检查
// @Description 返回各依赖状态
// @Tags base
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *BaseHandler) Health(c *gin.Context) {
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
		c.JSON(http.StatusServiceUnavailable, out)
		return
	}
	out["redis"] = "ok"
	if out["db"] != "ok" && out["db"] != "skip" {
		c.JSON(http.StatusServiceUnavailable, out)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Ping 简单存活探测。
// @Summary Ping
// @Tags base
// @Success 200 {object} response.Body
// @Router /api/v1/client/ping [get]
func (h *BaseHandler) Ping(c *gin.Context) {
	response.OK(c, gin.H{"pong": true})
}
