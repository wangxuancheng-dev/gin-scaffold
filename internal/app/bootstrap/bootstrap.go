package bootstrap

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"

	"gin-scaffold/api/handler"
	adminhandler "gin-scaffold/api/handler/admin"
	clienthandler "gin-scaffold/api/handler/client"
	"gin-scaffold/config"
	"gin-scaffold/internal/app/platform"
	"gin-scaffold/internal/dao"
	"gin-scaffold/internal/job"
	jobhandler "gin-scaffold/internal/job/handler"
	"gin-scaffold/internal/job/scheduler"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/internal/pkg/snowflake"
	websocketpkg "gin-scaffold/internal/pkg/websocket"
	"gin-scaffold/internal/service"
	"gin-scaffold/internal/service/authz"
	"gin-scaffold/middleware"
	"gin-scaffold/pkg/db"
	"gin-scaffold/pkg/httpclient"
	"gin-scaffold/pkg/limiter"
	"gin-scaffold/pkg/logger"
	"gin-scaffold/pkg/redis"
	"gin-scaffold/pkg/storage"
	"gin-scaffold/pkg/tracer"
	"gin-scaffold/routes"
)

type ServerDeps struct {
	Cfg     *config.App
	Engine  *gin.Engine
	cleanup func(context.Context)
}

func (d *ServerDeps) Cleanup(ctx context.Context) {
	if d == nil || d.cleanup == nil {
		return
	}
	d.cleanup(ctx)
}

type WorkerDeps struct {
	Cfg     *config.App
	Server  *asynq.Server
	Mux     *asynq.ServeMux
	cleanup func(context.Context)
}

func (d *WorkerDeps) Cleanup(ctx context.Context) {
	if d == nil || d.cleanup == nil {
		return
	}
	d.cleanup(ctx)
}

func InitServer(env, profile string) (*ServerDeps, error) {
	cfg, err := config.Load(env, profile)
	if err != nil {
		return nil, err
	}
	if err := db.SyncProcessLocalToTimeZone(cfg.DB.TimeZone); err != nil {
		return nil, fmt.Errorf("time.Local (align with db.time_zone / TIME_ZONE): %w", err)
	}
	httpclient.InitDefault(cfg.Outbound)
	if err := logger.Init(&cfg.Log); err != nil {
		return nil, err
	}
	platform.Init(cfg)

	cleanups := make([]func(context.Context), 0, 4)
	cleanups = append(cleanups, func(context.Context) { logger.Sync() })

	traceShutdown, err := tracer.Init(context.Background(), &cfg.Trace)
	if err != nil {
		runCleanups(context.Background(), cleanups)
		return nil, err
	}
	cleanups = append(cleanups, func(ctx context.Context) { _ = traceShutdown(ctx) })

	gdb, err := db.Init(&cfg.DB)
	if err != nil {
		runCleanups(context.Background(), cleanups)
		return nil, fmt.Errorf("database: %w", err)
	}
	if err := redis.Init(&cfg.Redis); err != nil {
		runCleanups(context.Background(), cleanups)
		return nil, fmt.Errorf("redis: %w", err)
	}
	cleanups = append(cleanups, func(context.Context) { _ = redis.Close() })
	if err := snowflake.Init(cfg.Snowflake.Node); err != nil {
		runCleanups(context.Background(), cleanups)
		return nil, fmt.Errorf("snowflake: %w", err)
	}
	if cfg.Storage.Enabled {
		sp, storageErr := storage.NewFromConfig(&cfg.Storage)
		if storageErr != nil {
			runCleanups(context.Background(), cleanups)
			return nil, fmt.Errorf("storage: %w", storageErr)
		}
		storage.InitDefault(sp)
	}

	var q *job.Client
	var inspector = job.NewInspector(cfg.Asynq.RedisAddr, cfg.Asynq.RedisPassword, cfg.Asynq.RedisDB)
	if cfg.Asynq.RedisAddr != "" {
		q = job.NewClient(
			cfg.Asynq.RedisAddr,
			cfg.Asynq.RedisPassword,
			cfg.Asynq.RedisDB,
			cfg.Asynq.Queue,
			cfg.Asynq.MaxRetry,
			cfg.Asynq.TimeoutSec,
			cfg.Asynq.DedupWindowSec,
		)
		cleanups = append(cleanups, func(context.Context) { _ = q.Close() })
	}

	jm := jwtpkg.NewManager(&cfg.JWT)
	userDAO := dao.NewUserDAO(gdb)
	menuDAO := dao.NewMenuDAO(gdb)
	taskDAO := dao.NewScheduledTaskDAO(gdb)
	auditDAO := dao.NewAuditLogDAO(gdb)
	outboxDAO := dao.NewOutboxDAO(gdb)
	sysSettingDAO := dao.NewSystemSettingDAO(gdb)
	announcementDAO := dao.NewAnnouncementDAO(gdb)
	authzDAO := dao.NewAuthzDAO(gdb)
	middleware.SetPermissionChecker(authz.NewDBPermissionChecker(authzDAO, cfg.RBAC.SuperAdminUserID))
	middleware.SetSuperAdminUserID(cfg.RBAC.SuperAdminUserID)
	userSvc := service.NewUserService(userDAO, q, jm, cfg.RBAC.SuperAdminUserID, gdb, outboxDAO, cfg.Outbox)
	menuSvc := service.NewMenuService(menuDAO)
	taskSvc := service.NewScheduledTaskService(taskDAO, cfg.Scheduler)
	sysSettingSvc := service.NewSystemSettingService(sysSettingDAO)
	announcementSvc := service.NewAnnouncementService(announcementDAO)
	hub := websocketpkg.NewHub()
	wsSvc := service.NewWSService(hub)
	sseSvc := service.NewSSEService()

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

	stopTaskScheduler, notifyTaskScheduler, err := scheduler.StartTaskScheduler(taskSvc, cfg.Scheduler)
	if err != nil {
		runCleanups(context.Background(), cleanups)
		return nil, fmt.Errorf("task scheduler: %w", err)
	}
	taskSvc.SetOnChanged(notifyTaskScheduler)
	cleanups = append(cleanups, func(context.Context) { stopTaskScheduler() })

	stopOutboxDispatcher := service.NewOutboxDispatcher(outboxDAO, cfg.Outbox).Start()
	cleanups = append(cleanups, func(context.Context) { stopOutboxDispatcher() })

	var lim limiter.Backend = limiter.NewStore(cfg.Limiter.IPRPS, cfg.Limiter.IPBurst, cfg.Limiter.RouteRPS, cfg.Limiter.RouteBurst)
	if strings.ToLower(strings.TrimSpace(cfg.Limiter.Mode)) == "redis" {
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
		lim = limiter.NewRedisStore(prefix, ws, cfg.Limiter.IPRPS, cfg.Limiter.IPBurst, cfg.Limiter.RouteRPS, cfg.Limiter.RouteBurst)
	}

	engine, err := routes.Build(routes.Options{
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
		Limiter:           lim,
	})
	if err != nil {
		runCleanups(context.Background(), cleanups)
		return nil, fmt.Errorf("routes: %w", err)
	}

	return &ServerDeps{
		Cfg:    cfg,
		Engine: engine,
		cleanup: func(ctx context.Context) {
			runCleanups(ctx, cleanups)
		},
	}, nil
}

func InitWorker(env, profile string) (*WorkerDeps, error) {
	cfg, err := config.Load(env, profile)
	if err != nil {
		return nil, err
	}
	if err := db.SyncProcessLocalToTimeZone(cfg.DB.TimeZone); err != nil {
		return nil, fmt.Errorf("time.Local (align with db.time_zone / TIME_ZONE): %w", err)
	}
	httpclient.InitDefault(cfg.Outbound)
	if cfg.Storage.Enabled {
		sp, storageErr := storage.NewFromConfig(&cfg.Storage)
		if storageErr != nil {
			return nil, fmt.Errorf("storage: %w", storageErr)
		}
		storage.InitDefault(sp)
	}
	if err := logger.Init(&cfg.Log); err != nil {
		return nil, err
	}
	platform.Init(cfg)
	if _, err := db.Init(&cfg.DB); err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}
	if err := redis.Init(&cfg.Redis); err != nil {
		return nil, fmt.Errorf("redis: %w", err)
	}

	cleanups := []func(context.Context){
		func(context.Context) { logger.Sync() },
		func(context.Context) { _ = redis.Close() },
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

	return &WorkerDeps{
		Cfg:    cfg,
		Server: srv,
		Mux:    mux,
		cleanup: func(ctx context.Context) {
			runCleanups(ctx, cleanups)
		},
	}, nil
}

func runCleanups(ctx context.Context, funcs []func(context.Context)) {
	for i := len(funcs) - 1; i >= 0; i-- {
		funcs[i](ctx)
	}
}

func resolveAsynqQueues(cfg config.AsynqConfig) map[string]int {
	if len(cfg.Queues) > 0 {
		return cfg.Queues
	}
	queue := cfg.Queue
	if queue == "" {
		queue = "default"
	}
	return map[string]int{queue: 1}
}
