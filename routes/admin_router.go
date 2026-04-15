package routes

import (
	"github.com/gin-gonic/gin"

	adminhandler "gin-scaffold/api/handler/admin"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/middleware"
)

func registerAdminRoutes(r *gin.Engine, jwtMgr *jwtpkg.Manager, user *adminhandler.UserHandler, menu *adminhandler.MenuHandler, ops *adminhandler.OpsHandler) {
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
	admin.GET("/users/export", middleware.RequirePermission("user:export"), user.Export)
	admin.GET("/menus", middleware.RequirePermission("menu:read"), menu.ListMine)
	admin.GET("/dbping", middleware.RequirePermission("db:ping"), ops.DBPing)
}
