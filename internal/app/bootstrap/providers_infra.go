package bootstrap

import (
	"strings"

	"github.com/hibiken/asynq"

	"gin-scaffold/internal/config"
	"gin-scaffold/internal/job"
	jobhandler "gin-scaffold/internal/job/handler"
	"gin-scaffold/pkg/limiter"
)

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
