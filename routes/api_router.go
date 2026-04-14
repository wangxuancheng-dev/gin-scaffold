package routes

import (
	"github.com/gin-gonic/gin"

	"gin-scaffold/api/handler"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
)

// registerAPIV1 注册 /api/v1 下客户端与后台路由。
func registerAPIV1(r *gin.Engine, jwtMgr *jwtpkg.Manager, base *handler.BaseHandler, user *handler.UserHandler, ws *handler.WSHandler, sse *handler.SSEHandler) {
	registerClientRoutes(r, jwtMgr, base, user, ws, sse)
	registerAdminRoutes(r, jwtMgr)
}
