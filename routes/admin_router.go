package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/middleware"
	"gin-scaffold/pkg/db"
)

func registerAdminRoutes(r *gin.Engine, jwtMgr *jwtpkg.Manager) {
	if jwtMgr == nil {
		return
	}

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
