package routes

import (
	"github.com/gin-gonic/gin"

	adminhandler "gin-scaffold/api/handler/admin"
	"gin-scaffold/api/handler"
	clienthandler "gin-scaffold/api/handler/client"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
)

// registerAPIV1 注册 /api/v1 下客户端与后台路由。
func registerAPIV1(
	r *gin.Engine,
	jwtMgr *jwtpkg.Manager,
	base *handler.BaseHandler,
	clientUser *clienthandler.UserHandler,
	adminUser *adminhandler.UserHandler,
	adminOps *adminhandler.OpsHandler,
	ws *handler.WSHandler,
	sse *handler.SSEHandler,
) {
	registerClientRoutes(r, jwtMgr, base, clientUser, ws, sse)
	registerAdminRoutes(r, jwtMgr, adminUser, adminOps)
}
