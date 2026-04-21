package adminroutes

import (
	"github.com/gin-gonic/gin"

	adminhandler "gin-scaffold/internal/api/handler/admin"
	"gin-scaffold/internal/middleware"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
)

func Register(r *gin.Engine, jwtMgr *jwtpkg.Manager, user *adminhandler.UserHandler, menu *adminhandler.MenuHandler, ops *adminhandler.OpsHandler, task *adminhandler.TaskHandler, sys *adminhandler.SystemSettingHandler, queue *adminhandler.TaskQueueHandler, generatedAnnouncement *adminhandler.AnnouncementHandler) {
	if jwtMgr == nil {
		return
	}

	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.JWTAuth(jwtMgr))
	admin.Use(middleware.RequireRoles("admin"))
	registerAdminUserRoutes(admin, user)
	registerAdminMenuRoutes(admin, menu)
	registerAdminOpsRoutes(admin, ops)
	registerAdminTaskRoutes(admin, task)
	registerAdminTaskQueueRoutes(admin, queue)
	registerAdminSystemSettingRoutes(admin, sys)
	registerAdminAnnouncementRoutes(admin, generatedAnnouncement)
}
