package bootstrap

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"

	"gin-scaffold/internal/app/platform"
	"gin-scaffold/internal/config"
	"gin-scaffold/internal/job"
	"gin-scaffold/internal/job/scheduler"
	"gin-scaffold/internal/pkg/snowflake"
	"gin-scaffold/internal/routes"
	"gin-scaffold/internal/service"
	"gin-scaffold/pkg/db"
	"gin-scaffold/pkg/httpclient"
	"gin-scaffold/pkg/logger"
	"gin-scaffold/pkg/redis"
	"gin-scaffold/pkg/storage"
	"gin-scaffold/pkg/tracer"
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

	providers := buildServerProviders(cfg, gdb, q, inspector)

	stopTaskScheduler, notifyTaskScheduler, err := scheduler.StartTaskScheduler(providers.taskSvc, cfg.Scheduler)
	if err != nil {
		runCleanups(context.Background(), cleanups)
		return nil, fmt.Errorf("task scheduler: %w", err)
	}
	providers.taskSvc.SetOnChanged(notifyTaskScheduler)
	cleanups = append(cleanups, func(context.Context) { stopTaskScheduler() })

	stopOutboxDispatcher := service.NewOutboxDispatcher(providers.outboxDAO, cfg.Outbox).Start()
	cleanups = append(cleanups, func(context.Context) { stopOutboxDispatcher() })
	engine, err := routes.Build(providers.routeOpts)
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

	providers := buildWorkerProviders(cfg)

	return &WorkerDeps{
		Cfg:    cfg,
		Server: providers.server,
		Mux:    providers.mux,
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
