package adminroutes

import (
	"github.com/gin-gonic/gin"

	adminhandler "gin-scaffold/api/handler/admin"
	"gin-scaffold/middleware"
)

func registerAdminMenuRoutes(admin *gin.RouterGroup, h *adminhandler.MenuHandler) {
	// 具体路径需先于 /menus/:id 注册，避免被 id 参数吞掉。
	admin.GET("/menus/catalog", middleware.RequirePermission("menu:catalog"), h.Catalog)
	admin.GET("/menus", middleware.RequirePermission("menu:read"), h.ListMine)
	admin.POST("/menus", middleware.RequirePermission("menu:create"), h.Create)
	admin.GET("/menus/:id", middleware.RequirePermission("menu:catalog"), h.Get)
	admin.PUT("/menus/:id", middleware.RequirePermission("menu:update"), h.Update)
	admin.DELETE("/menus/:id", middleware.RequirePermission("menu:delete"), h.Delete)
}
