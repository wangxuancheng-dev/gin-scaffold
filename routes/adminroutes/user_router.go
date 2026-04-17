package adminroutes

import (
	"github.com/gin-gonic/gin"

	adminhandler "gin-scaffold/api/handler/admin"
	"gin-scaffold/middleware"
)

func registerAdminUserRoutes(admin *gin.RouterGroup, h *adminhandler.UserHandler) {
	admin.GET("/users", middleware.RequirePermission("user:read"), h.List)
	admin.GET("/users/:id", middleware.RequirePermission("user:read"), h.Get)
	admin.POST("/users", middleware.RequirePermission("user:create"), h.Create)
	admin.PUT("/users/:id", middleware.RequirePermission("user:update"), h.Update)
	admin.DELETE("/users/:id", middleware.RequirePermission("user:delete"), h.Delete)
	admin.POST("/users/export/tasks", middleware.RequirePermission("user:export"), h.ExportTaskCreate)
	admin.GET("/users/export/tasks/:task_id", middleware.RequirePermission("user:export"), h.ExportTaskStatus)
	admin.GET("/users/export/tasks/:task_id/download", middleware.RequirePermission("user:export"), h.ExportTaskDownload)
}
