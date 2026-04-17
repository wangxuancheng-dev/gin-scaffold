package routes

import (
	"github.com/gin-gonic/gin"

	"gin-scaffold/api/handler"
	clienthandler "gin-scaffold/api/handler/client"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/middleware"
)

func registerClientRoutes(r *gin.Engine, jwtMgr *jwtpkg.Manager, base *handler.BaseHandler, user *clienthandler.UserHandler, file *clienthandler.FileHandler, ws *handler.WSHandler, sse *handler.SSEHandler) {
	if file == nil {
		file = clienthandler.NewFileHandler(nil)
	}
	client := r.Group("/api/v1/client")
	{
		client.GET("/ping", base.Ping)
		client.POST("/users", user.Register)
		client.POST("/auth/login", user.Login)
		client.POST("/auth/refresh", user.Refresh)
		client.GET("/files/*key/download", file.Download)
		client.GET("/ws", ws.Handle)
		client.GET("/sse/stream", sse.Stream)
	}
	if jwtMgr == nil {
		return
	}
	clientAuth := r.Group("/api/v1/client")
	clientAuth.Use(middleware.JWTAuth(jwtMgr))
	clientAuth.GET("/users/:id", user.Get)
	clientAuth.POST("/auth/logout", user.Logout)
	clientAuth.POST("/files/upload", file.Upload)
	clientAuth.POST("/files/presign", file.PresignPut)
	clientAuth.GET("/files/*key/url", file.SignURL)
}
