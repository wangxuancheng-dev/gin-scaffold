// Package routes 注册 Gin 路由、全局中间件与 Swagger。
package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	adminhandler "gin-scaffold/api/handler/admin"
	"gin-scaffold/api/handler"
	clienthandler "gin-scaffold/api/handler/client"
	"gin-scaffold/config"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/middleware"
)

// Options 路由依赖。
type Options struct {
	Cfg        *config.App
	JWT        *jwtpkg.Manager
	Base       *handler.BaseHandler
	ClientUser *clienthandler.UserHandler
	AdminUser  *adminhandler.UserHandler
	WS         *handler.WSHandler
	SSE        *handler.SSEHandler
	TraceOn    bool
}

// Build 构建 *gin.Engine。
func Build(opts Options) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery())
	r.Use(middleware.AccessLog())
	r.Use(middleware.I18n())
	r.Use(middleware.CORS(&opts.Cfg.CORS))
	if opts.Cfg != nil {
		r.Use(middleware.Limiter(
			opts.Cfg.Limiter.IPRPS,
			opts.Cfg.Limiter.IPBurst,
			opts.Cfg.Limiter.RouteRPS,
			opts.Cfg.Limiter.RouteBurst,
		))
	}
	if opts.TraceOn {
		r.Use(otelgin.Middleware(opts.Cfg.Name))
	}
	if opts.Cfg != nil && opts.Cfg.Metrics.Enabled {
		middleware.Metrics(r, "gin")
	}

	r.GET("/health", opts.Base.Health)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	registerAPIV1(r, opts.JWT, opts.Base, opts.ClientUser, opts.AdminUser, opts.WS, opts.SSE)

	return r
}
