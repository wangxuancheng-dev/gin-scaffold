package adminroutes

import (
	"github.com/gin-gonic/gin"

	adminhandler "gin-scaffold/api/handler/admin"
	"gin-scaffold/middleware"
)

func registerAdminSystemSettingRoutes(admin *gin.RouterGroup, h *adminhandler.SystemSettingHandler) {
	admin.GET("/system-settings", middleware.RequirePermission("sys:config:read"), h.List)
	admin.GET("/system-settings/:id", middleware.RequirePermission("sys:config:read"), h.Get)
	admin.POST("/system-settings", middleware.RequirePermission("sys:config:write"), h.Create)
	admin.PUT("/system-settings/:id", middleware.RequirePermission("sys:config:write"), h.Update)
	admin.DELETE("/system-settings/:id", middleware.RequirePermission("sys:config:write"), h.Delete)
}
