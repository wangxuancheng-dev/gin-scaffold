package adminroutes

import (
	"github.com/gin-gonic/gin"

	adminhandler "gin-scaffold/api/handler/admin"
	"gin-scaffold/middleware"
)

func registerAdminTaskRoutes(admin *gin.RouterGroup, h *adminhandler.TaskHandler) {
	admin.GET("/tasks", middleware.RequirePermission("task:read"), h.List)
	admin.POST("/tasks", middleware.RequirePermission("task:create"), h.Create)
	admin.PUT("/tasks/:id", middleware.RequirePermission("task:update"), h.Update)
	admin.DELETE("/tasks/:id", middleware.RequirePermission("task:delete"), h.Delete)
	admin.POST("/tasks/:id/toggle", middleware.RequirePermission("task:toggle"), h.Toggle)
	admin.POST("/tasks/:id/run", middleware.RequirePermission("task:run"), h.RunNow)
	admin.GET("/tasks/:id/logs", middleware.RequirePermission("task:read"), h.Logs)
}
