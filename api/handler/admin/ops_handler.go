package adminhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gin-scaffold/api/response"
	"gin-scaffold/internal/pkg/errcode"
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
		response.FailHTTP(c, http.StatusServiceUnavailable, errcode.InternalError, errcode.KeyInternal, "db not configured")
		return
	}
	sqlDB, err := db.DB().DB()
	if err != nil {
		response.FailHTTP(c, http.StatusServiceUnavailable, errcode.InternalError, errcode.KeyInternal, err.Error())
		return
	}
	if err := sqlDB.PingContext(c.Request.Context()); err != nil {
		response.FailHTTP(c, http.StatusServiceUnavailable, errcode.InternalError, errcode.KeyInternal, err.Error())
		return
	}
	response.OK(c, gin.H{"db": "ok"})
}
