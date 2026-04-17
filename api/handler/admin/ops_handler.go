package adminhandler

import (
	"github.com/gin-gonic/gin"

	"gin-scaffold/api/handler"
	"gin-scaffold/api/response"
	"gin-scaffold/pkg/db"
)

// OpsHandler 后台运维类接口。
type OpsHandler struct{}

// NewOpsHandler 构造后台运维 handler。
func NewOpsHandler() *OpsHandler {
	return &OpsHandler{}
}

// DBPing 检查数据库连通性。
// @Summary 数据库连通性检查（后台）
// @Tags admin-ops
// @Produce json
// @Success 200 {object} response.Body
// @Router /api/v1/admin/dbping [get]
func (h *OpsHandler) DBPing(c *gin.Context) {
	if db.DB() == nil {
		handler.FailServiceUnavailable(c, nil, "db not configured")
		return
	}
	sqlDB, err := db.DB().DB()
	if err != nil {
		handler.FailServiceUnavailable(c, err, "")
		return
	}
	if err := sqlDB.PingContext(c.Request.Context()); err != nil {
		handler.FailServiceUnavailable(c, err, "")
		return
	}
	response.OK(c, gin.H{"db": "ok"})
}
