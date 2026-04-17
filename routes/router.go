// Package routes 注册 Gin 路由、全局中间件与 Swagger。
package routes

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"gin-scaffold/api/handler"
	adminhandler "gin-scaffold/api/handler/admin"
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
	AdminMenu  *adminhandler.MenuHandler
	AdminOps   *adminhandler.OpsHandler
	AdminTask  *adminhandler.TaskHandler
	WS         *handler.WSHandler
	SSE        *handler.SSEHandler
	TraceOn    bool
}

// Build 构建 *gin.Engine。
func Build(opts Options) *gin.Engine {
	if opts.Cfg != nil && opts.Cfg.Debug {
		gin.SetMode(gin.DebugMode)
		gin.ForceConsoleColor()
		gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
			fmt.Fprintf(os.Stdout, "\x1b[36m[ROUTE]\x1b[0m %-6s %-40s \x1b[90mhandlers=%-2d\x1b[0m %s\n",
				httpMethod,
				truncatePath(absolutePath, 40),
				nuHandlers,
				handlerName,
			)
		}
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.RequestID())
	if opts.Cfg != nil {
		r.Use(middleware.BodyLimit(opts.Cfg.HTTP.MaxBodyBytes))
	}
	if opts.Cfg != nil && opts.Cfg.Debug {
		// Debug 模式下额外输出 Gin 风格请求日志，便于本地排障。
		r.Use(gin.Logger())
	}
	r.Use(middleware.Recovery(opts.Cfg != nil && opts.Cfg.Debug))
	r.Use(middleware.AccessLog())
	r.Use(middleware.I18n(&opts.Cfg.I18n))
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
		middleware.Metrics(r, "gin", opts.Cfg.Metrics.Path)
	}

	r.GET("/livez", opts.Base.Livez)
	r.GET("/readyz", opts.Base.Readyz)
	r.GET("/health", opts.Base.Health)
	if opts.Cfg != nil && opts.Cfg.Debug {
		// 仅调试环境：用于验证 Recovery 与日志链路。
		r.GET("/debug/panic", func(c *gin.Context) {
			if !isLoopbackClient(c.ClientIP()) {
				c.AbortWithStatus(404)
				return
			}
			panic("debug panic endpoint")
		})
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	registerAPIV1(r, opts.JWT, opts.Base, opts.ClientUser, opts.AdminUser, opts.AdminMenu, opts.AdminOps, opts.AdminTask, opts.WS, opts.SSE)

	return r
}

func truncatePath(p string, max int) string {
	if len(p) <= max {
		return p + strings.Repeat(" ", max-len(p))
	}
	if max <= 3 {
		return p[:max]
	}
	return p[:max-3] + "..."
}

func isLoopbackClient(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil && parsed.IsLoopback()
}
