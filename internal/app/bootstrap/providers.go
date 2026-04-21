package bootstrap

import (
	"strings"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"

	"gin-scaffold/internal/api/handler"
	adminhandler "gin-scaffold/internal/api/handler/admin"
	clienthandler "gin-scaffold/internal/api/handler/client"
	"gin-scaffold/internal/config"
	"gin-scaffold/internal/dao"
	"gin-scaffold/internal/job"
	jobhandler "gin-scaffold/internal/job/handler"
	"gin-scaffold/internal/middleware"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	websocketpkg "gin-scaffold/internal/pkg/websocket"
	"gin-scaffold/internal/routes"
	"gin-scaffold/internal/service"
	"gin-scaffold/internal/service/authz"
	"gin-scaffold/pkg/limiter"
)

type serverProviders struct {
	taskSvc   *service.ScheduledTaskService
	outboxDAO *dao.OutboxDAO
	routeOpts routes.Options
}

type workerProviders struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

func buildServerProviders(cfg *config.App, gdb *gorm.DB, q *job.Client, inspector *asynq.Inspector) *serverProviders {
	jm := jwtpkg.NewManager(&cfg.JWT)

	// Data providers.
	userDAO := dao.NewUserDAO(gdb)
	menuDAO := dao.NewMenuDAO(gdb)
	taskDAO := dao.NewScheduledTaskDAO(gdb)
	auditDAO := dao.NewAuditLogDAO(gdb)
	outboxDAO := dao.NewOutboxDAO(gdb)
	sysSettingDAO := dao.NewSystemSettingDAO(gdb)
	announcementDAO := dao.NewAnnouncementDAO(gdb)
	authzDAO := dao.NewAuthzDAO(gdb)

	// Infra providers.
	middleware.SetPermissionChecker(authz.NewDBPermissionChecker(authzDAO, cfg.RBAC.SuperAdminUserID))
	middleware.SetSuperAdminUserID(cfg.RBAC.SuperAdminUserID)

	// Service providers.
	userSvc := service.NewUserService(userDAO, q, jm, cfg.RBAC.SuperAdminUserID, gdb, outboxDAO, cfg.Outbox)
	menuSvc := service.NewMenuService(menuDAO)
	taskSvc := service.NewScheduledTaskService(taskDAO, cfg.Scheduler)
	sysSettingSvc := service.NewSystemSettingService(sysSettingDAO)
	announcementSvc := service.NewAnnouncementService(announcementDAO)
	hub := websocketpkg.NewHub()
	wsSvc := service.NewWSService(hub)
	sseSvc := service.NewSSEService()

	// HTTP handler providers.
	baseH := &handler.BaseHandler{DB: gdb, Storage: &cfg.Storage}
	clientUserH := clienthandler.NewUserHandler(userSvc)
	clientFileH := clienthandler.NewFileHandler(&cfg.Storage)
	adminUserH := adminhandler.NewUserHandler(userSvc, q)
	adminMenuH := adminhandler.NewMenuHandler(menuSvc)
	adminOpsH := adminhandler.NewOpsHandler(auditDAO, q)
	adminTaskH := adminhandler.NewTaskHandler(taskSvc)
	adminSysH := adminhandler.NewSystemSettingHandler(sysSettingSvc)
	adminQueueH := adminhandler.NewTaskQueueHandler(inspector)
	adminAnnouncementH := adminhandler.NewAnnouncementHandler(announcementSvc)
	wsH := handler.NewWSHandler(wsSvc, middleware.WebSocketCheckOrigin(cfg.CORS.AllowOrigins))
	sseH := handler.NewSSEHandler(sseSvc)

	return &serverProviders{
		taskSvc:   taskSvc,
		outboxDAO: outboxDAO,
		routeOpts: routes.Options{
			Cfg:               cfg,
			JWT:               jm,
			Base:              baseH,
			ClientUser:        clientUserH,
			ClientFile:        clientFileH,
			AdminUser:         adminUserH,
			AdminMenu:         adminMenuH,
			AdminOps:          adminOpsH,
			AdminTask:         adminTaskH,
			AdminSys:          adminSysH,
			AdminQueue:        adminQueueH,
			AdminAnnouncement: adminAnnouncementH,
			WS:                wsH,
			SSE:               sseH,
			TraceOn:           cfg.Trace.Enabled,
			Limiter:           buildLimiterProvider(cfg),
		},
	}
}

func buildLimiterProvider(cfg *config.App) limiter.Backend {
	if cfg == nil {
		return nil
	}
	lim := limiter.NewStoreWithOptions(limiter.StoreOptions{
		WindowSec:         cfg.Limiter.WindowSec,
		IPMaxPerWindow:    cfg.Limiter.IPMaxPerWindow,
		RouteMaxPerWindow: cfg.Limiter.RouteMaxPerWindow,
		IPRPS:             cfg.Limiter.IPRPS,
		IPBurst:           cfg.Limiter.IPBurst,
		RouteRPS:          cfg.Limiter.RouteRPS,
		RouteBurst:        cfg.Limiter.RouteBurst,
	})
	if strings.ToLower(strings.TrimSpace(cfg.Limiter.Mode)) != "redis" {
		return lim
	}
	ws := cfg.Limiter.WindowSec
	if ws <= 0 {
		ws = 1
	}
	prefix := strings.TrimSpace(cfg.Limiter.RedisKeyPrefix)
	if prefix == "" {
		prefix = strings.TrimSpace(cfg.Platform.Cache.KeyPrefix)
	}
	if prefix != "" && !strings.HasSuffix(prefix, ":") {
		prefix += ":"
	}
	if prefix == "" {
		prefix = "app:"
	}
	return limiter.NewRedisStore(
		prefix,
		ws,
		cfg.Limiter.IPRPS,
		cfg.Limiter.IPBurst,
		cfg.Limiter.RouteRPS,
		cfg.Limiter.RouteBurst,
		cfg.Limiter.IPMaxPerWindow,
		cfg.Limiter.RouteMaxPerWindow,
	)
}

func buildWorkerProviders(cfg *config.App) *workerProviders {
	if cfg == nil {
		return &workerProviders{}
	}
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Asynq.RedisAddr,
			Password: cfg.Asynq.RedisPassword,
			DB:       cfg.Asynq.RedisDB,
		},
		asynq.Config{
			Concurrency:    cfg.Asynq.Concurrency,
			StrictPriority: cfg.Asynq.StrictPriority,
			Queues:         resolveAsynqQueues(cfg.Asynq),
		},
	)
	mux := asynq.NewServeMux()
	mux.Handle(job.TypeWelcomeEmail, jobhandler.WelcomeHandler{})
	mux.Handle(job.TypeAuditExport, jobhandler.AuditExportHandler{})
	mux.Handle(job.TypeUserExport, jobhandler.UserExportHandler{})
	return &workerProviders{server: srv, mux: mux}
}
