package adminroutes

import (
	"github.com/gin-gonic/gin"

	adminhandler "gin-scaffold/api/handler/admin"
	"gin-scaffold/middleware"
)

func registerAdminTaskQueueRoutes(admin *gin.RouterGroup, h *adminhandler.TaskQueueHandler) {
	admin.GET("/task-queues/summary", middleware.RequirePermission("task:read"), h.Summary)
	admin.GET("/task-queues/failed", middleware.RequirePermission("task:read"), h.FailedList)
	admin.POST("/task-queues/:queue/failed/:task_id/retry", middleware.RequirePermission("task:run"), h.Retry)
	admin.POST("/task-queues/:queue/failed/:task_id/archive", middleware.RequirePermission("task:update"), h.Archive)
}
