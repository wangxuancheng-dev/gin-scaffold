package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gin-scaffold/api/handler"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/middleware"
	"gin-scaffold/pkg/db"
)

// registerAPIV1 注册 /api/v1 业务路由与需鉴权子路由。
func registerAPIV1(r *gin.Engine, jwtMgr *jwtpkg.Manager, base *handler.BaseHandler, user *handler.UserHandler, ws *handler.WSHandler, sse *handler.SSEHandler) {
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", base.Ping)
		v1.POST("/users", user.Register)
		v1.GET("/users", user.List)
		v1.POST("/auth/login", user.Login)
		v1.POST("/auth/refresh", user.Refresh)
		v1.GET("/ws", ws.Handle)
		v1.GET("/sse/stream", sse.Stream)
	}
	if jwtMgr != nil {
		authz := r.Group("/api/v1")
		authz.Use(middleware.JWTAuth(jwtMgr))
		authz.GET("/users/:id", user.Get)
		authz.POST("/auth/logout", user.Logout)

		admin := r.Group("/api/v1/admin")
		admin.Use(middleware.JWTAuth(jwtMgr))
		admin.Use(middleware.RequireRoles("admin"))
		admin.Use(middleware.RequirePermission("db:ping"))
		admin.GET("/dbping", func(c *gin.Context) {
			if db.DB() == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"db": "not configured"})
				return
			}
			sqlDB, err := db.DB().DB()
			if err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"err": err.Error()})
				return
			}
			if err := sqlDB.PingContext(c.Request.Context()); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"err": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"db": "ok"})
		})
	}
}
