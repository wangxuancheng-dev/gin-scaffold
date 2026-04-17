package routes

import (
	"github.com/gin-gonic/gin"

	"gin-scaffold/api/handler"
	adminhandler "gin-scaffold/api/handler/admin"
	clienthandler "gin-scaffold/api/handler/client"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
)

// registerAPIV1 注册 /api/v1 下客户端与后台路由。
func registerAPIV1(
	r *gin.Engine,
	jwtMgr *jwtpkg.Manager,
	base *handler.BaseHandler,
	clientUser *clienthandler.UserHandler,
	clientFile *clienthandler.FileHandler,
	adminUser *adminhandler.UserHandler,
	adminMenu *adminhandler.MenuHandler,
	adminOps *adminhandler.OpsHandler,
	adminTask *adminhandler.TaskHandler,
	adminSys *adminhandler.SystemSettingHandler,
	adminQueue *adminhandler.TaskQueueHandler,
	adminAnnouncement *adminhandler.AnnouncementHandler,
	ws *handler.WSHandler,
	sse *handler.SSEHandler,
) {
	registerClientRoutes(r, jwtMgr, base, clientUser, clientFile, ws, sse)
	registerAdminRoutes(r, jwtMgr, adminUser, adminMenu, adminOps, adminTask, adminSys, adminQueue, adminAnnouncement)
}
