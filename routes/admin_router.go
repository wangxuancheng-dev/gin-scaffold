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
	admin.GET("/users", middleware.RequirePermission("user:rw"), user.List)
	admin.GET("/users/export", middleware.RequirePermission("user:rw"), user.Export)
	admin.GET("/menus", middleware.RequirePermission("menu:read"), menu.ListMine)
	admin.GET("/dbping", middleware.RequirePermission("db:ping"), ops.DBPing)
}
