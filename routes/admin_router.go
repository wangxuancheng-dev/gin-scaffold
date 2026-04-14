package routes

import (
	"github.com/gin-gonic/gin"

	adminhandler "gin-scaffold/api/handler/admin"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/middleware"
)

func registerAdminRoutes(r *gin.Engine, jwtMgr *jwtpkg.Manager, user *adminhandler.UserHandler, ops *adminhandler.OpsHandler) {
	if jwtMgr == nil {
		return
	}

	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.JWTAuth(jwtMgr))
	admin.Use(middleware.RequireRoles("admin"))
	admin.Use(middleware.RequirePermission("db:ping"))
	admin.GET("/users", user.List)
	admin.GET("/dbping", ops.DBPing)
}
