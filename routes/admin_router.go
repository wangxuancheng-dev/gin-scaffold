package routes

import (
	"github.com/gin-gonic/gin"

	adminhandler "gin-scaffold/api/handler/admin"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/middleware"
)

func registerAdminRoutes(r *gin.Engine, jwtMgr *jwtpkg.Manager, user *adminhandler.UserHandler, menu *adminhandler.MenuHandler, ops *adminhandler.OpsHandler, task *adminhandler.TaskHandler, sys *adminhandler.SystemSettingHandler, queue *adminhandler.TaskQueueHandler, generatedAnnouncement *adminhandler.AnnouncementHandler) {
	if jwtMgr == nil {
		return
	}

	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.JWTAuth(jwtMgr))
	admin.Use(middleware.RequireRoles("admin"))
	admin.GET("/users", middleware.RequirePermission("user:read"), user.List)
	admin.GET("/users/:id", middleware.RequirePermission("user:read"), user.Get)
	admin.POST("/users", middleware.RequirePermission("user:create"), user.Create)
	admin.PUT("/users/:id", middleware.RequirePermission("user:update"), user.Update)
	admin.DELETE("/users/:id", middleware.RequirePermission("user:delete"), user.Delete)
	admin.POST("/users/export/tasks", middleware.RequirePermission("user:export"), user.ExportTaskCreate)
	admin.GET("/users/export/tasks/:task_id", middleware.RequirePermission("user:export"), user.ExportTaskStatus)
	admin.GET("/users/export/tasks/:task_id/download", middleware.RequirePermission("user:export"), user.ExportTaskDownload)
	admin.GET("/menus", middleware.RequirePermission("menu:read"), menu.ListMine)
	admin.GET("/dbping", middleware.RequirePermission("db:ping"), ops.DBPing)
	admin.GET("/audit-logs", middleware.RequirePermission("audit:read"), ops.AuditLogs)
	admin.POST("/audit-logs/export/tasks", middleware.RequirePermission("audit:export"), ops.AuditLogsExportTaskCreate)
	admin.GET("/audit-logs/export/tasks/:task_id", middleware.RequirePermission("audit:export"), ops.AuditLogsExportTaskStatus)
	admin.GET("/audit-logs/export/tasks/:task_id/download", middleware.RequirePermission("audit:export"), ops.AuditLogsExportTaskDownload)
	admin.GET("/tasks", middleware.RequirePermission("task:read"), task.List)
	admin.POST("/tasks", middleware.RequirePermission("task:create"), task.Create)
	admin.PUT("/tasks/:id", middleware.RequirePermission("task:update"), task.Update)
	admin.DELETE("/tasks/:id", middleware.RequirePermission("task:delete"), task.Delete)
	admin.POST("/tasks/:id/toggle", middleware.RequirePermission("task:toggle"), task.Toggle)
	admin.POST("/tasks/:id/run", middleware.RequirePermission("task:run"), task.RunNow)
	admin.GET("/tasks/:id/logs", middleware.RequirePermission("task:read"), task.Logs)
	admin.GET("/task-queues/summary", middleware.RequirePermission("task:read"), queue.Summary)
	admin.GET("/task-queues/failed", middleware.RequirePermission("task:read"), queue.FailedList)
	admin.POST("/task-queues/:queue/failed/:task_id/retry", middleware.RequirePermission("task:run"), queue.Retry)
	admin.POST("/task-queues/:queue/failed/:task_id/archive", middleware.RequirePermission("task:update"), queue.Archive)
	admin.GET("/system-settings", middleware.RequirePermission("sys:config:read"), sys.List)
	admin.GET("/system-settings/:id", middleware.RequirePermission("sys:config:read"), sys.Get)
	admin.POST("/system-settings", middleware.RequirePermission("sys:config:write"), sys.Create)
	admin.PUT("/system-settings/:id", middleware.RequirePermission("sys:config:write"), sys.Update)
	admin.DELETE("/system-settings/:id", middleware.RequirePermission("sys:config:write"), sys.Delete)
	registerAdminAnnouncementRoutes(admin, generatedAnnouncement)
}
