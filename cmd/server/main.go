// Package main 为服务入口：配置加载、组件初始化、HTTP 服务与 Asynq Worker 子命令。
//
// @title           Gin Scaffold API
// @version         1.0
// @description     企业级 Gin 脚手架示例 API
// @BasePath        /
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	cli "github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"gin-scaffold/api/handler"
	"gin-scaffold/config"
	"gin-scaffold/internal/dao"
	"gin-scaffold/internal/job"
	jobhandler "gin-scaffold/internal/job/handler"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/internal/pkg/snowflake"
	websocketpkg "gin-scaffold/internal/pkg/websocket"
	"gin-scaffold/internal/service"
	"gin-scaffold/pkg/db"
	"gin-scaffold/pkg/logger"
	"gin-scaffold/pkg/redis"
	"gin-scaffold/pkg/tracer"
	"gin-scaffold/routes"

	_ "gin-scaffold/docs"
)

func main() {
	app := &cli.App{
		Name:  "gin-scaffold",
		Usage: "HTTP API 或 Asynq Worker",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "env", Value: "dev", Usage: "配置环境: dev|test|prod"},
			&cli.StringFlag{Name: "profile", Value: "", Usage: "配置画像: 多实例标识，如 order/crm"},
		},
		Commands: []*cli.Command{
			{
				Name:  "server",
				Usage: "启动 HTTP 服务",
				Action: func(c *cli.Context) error {
					return runServer(c.String("env"), c.String("profile"))
				},
			},
			{
				Name:  "worker",
				Usage: "启动 Asynq 任务消费者",
				Action: func(c *cli.Context) error {
					return runWorker(c.String("env"), c.String("profile"))
				},
			},
		},
		Action: func(c *cli.Context) error {
			return runServer(c.String("env"), c.String("profile"))
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runServer(env, profile string) error {
	cfg, err := config.Load(env, profile)
	if err != nil {
		return err
	}
	if err := logger.Init(&cfg.Log); err != nil {
		return err
	}
	defer logger.Sync()
	printConfigSource("server")

	traceShutdown, err := tracer.Init(context.Background(), &cfg.Trace)
	if err != nil {
		return err
	}
	defer func() { _ = traceShutdown(context.Background()) }()

	gdb, err := db.Init(&cfg.DB)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	if err := redis.Init(&cfg.Redis); err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	if err := snowflake.Init(cfg.Snowflake.Node); err != nil {
		return fmt.Errorf("snowflake: %w", err)
	}

	var q *job.Client
	if cfg.Asynq.RedisAddr != "" {
		q = job.NewClient(cfg.Asynq.RedisAddr, cfg.Asynq.RedisPassword, cfg.Asynq.RedisDB)
		defer func() { _ = q.Close() }()
	}

	jm := jwtpkg.NewManager(&cfg.JWT)

	userDAO := dao.NewUserDAO(gdb)
	userSvc := service.NewUserService(userDAO, q, jm)
	hub := websocketpkg.NewHub()
	wsSvc := service.NewWSService(hub)
	sseSvc := service.NewSSEService()

	baseH := &handler.BaseHandler{DB: gdb}
	userH := handler.NewUserHandler(userSvc)
	wsH := handler.NewWSHandler(wsSvc)
	sseH := handler.NewSSEHandler(sseSvc)

	engine := routes.Build(routes.Options{
		Cfg:     cfg,
		JWT:     jm,
		Base:    baseH,
		User:    userH,
		WS:      wsH,
		SSE:     sseH,
		TraceOn: cfg.Trace.Enabled,
	})

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  time.Duration(cfg.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.HTTP.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.HTTP.IdleTimeout) * time.Second,
	}

	go func() {
		logger.InfoX("http listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.ErrorX("http server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.InfoX("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

func runWorker(env, profile string) error {
	cfg, err := config.Load(env, profile)
	if err != nil {
		return err
	}
	if err := logger.Init(&cfg.Log); err != nil {
		return err
	}
	defer logger.Sync()
	printConfigSource("worker")

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Asynq.RedisAddr,
			Password: cfg.Asynq.RedisPassword,
			DB:       cfg.Asynq.RedisDB,
		},
		asynq.Config{
			Concurrency:    cfg.Asynq.Concurrency,
			StrictPriority: cfg.Asynq.StrictPriority,
		},
	)
	mux := asynq.NewServeMux()
	mux.Handle(job.TypeWelcomeEmail, jobhandler.WelcomeHandler{})

	go func() {
		logger.InfoX("asynq worker started", zap.String("redis", cfg.Asynq.RedisAddr))
		if err := srv.Run(mux); err != nil {
			logger.ErrorX("asynq server error", zap.Error(err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	srv.Shutdown()
	return nil
}

func printConfigSource(component string) {
	src := config.Source()
	logger.InfoX(
		"config source summary",
		zap.String("component", component),
		zap.String("env", src.Env),
		zap.String("profile", src.Profile),
		zap.Strings("yaml_files", src.YAMLFiles),
		zap.Strings("dotenv_files", src.DotEnvFiles),
		zap.String("env_strategy", "runtime env vars have highest priority"),
	)
}
