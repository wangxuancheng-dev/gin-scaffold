package adminroutes

import (
	"github.com/gin-gonic/gin"

	adminhandler "gin-scaffold/api/handler/admin"
	"gin-scaffold/middleware"
)

func registerAdminOpsRoutes(admin *gin.RouterGroup, h *adminhandler.OpsHandler) {
	admin.GET("/dbping", middleware.RequirePermission("db:ping"), h.DBPing)
	admin.GET("/audit-logs", middleware.RequirePermission("audit:read"), h.AuditLogs)
	admin.POST("/audit-logs/export/tasks", middleware.RequirePermission("audit:export"), h.AuditLogsExportTaskCreate)
	admin.GET("/audit-logs/export/tasks/:task_id", middleware.RequirePermission("audit:export"), h.AuditLogsExportTaskStatus)
	admin.GET("/audit-logs/export/tasks/:task_id/download", middleware.RequirePermission("audit:export"), h.AuditLogsExportTaskDownload)
}
