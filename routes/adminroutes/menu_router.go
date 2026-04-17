package adminroutes

import (
	"github.com/gin-gonic/gin"

	adminhandler "gin-scaffold/api/handler/admin"
	"gin-scaffold/middleware"
)

func registerAdminMenuRoutes(admin *gin.RouterGroup, h *adminhandler.MenuHandler) {
	admin.GET("/menus", middleware.RequirePermission("menu:read"), h.ListMine)
}
