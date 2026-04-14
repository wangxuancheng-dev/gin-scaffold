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

	cli "github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"gin-scaffold/config"
	"gin-scaffold/internal/app/bootstrap"
	"gin-scaffold/pkg/logger"

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
	deps, err := bootstrap.InitServer(env, profile)
	if err != nil {
		return err
	}
	defer deps.Cleanup(context.Background())
	printConfigSource("server")
	addr := fmt.Sprintf("%s:%d", deps.Cfg.HTTP.Host, deps.Cfg.HTTP.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      deps.Engine,
		ReadTimeout:  time.Duration(deps.Cfg.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(deps.Cfg.HTTP.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(deps.Cfg.HTTP.IdleTimeout) * time.Second,
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
	deps, err := bootstrap.InitWorker(env, profile)
	if err != nil {
		return err
	}
	defer deps.Cleanup(context.Background())
	printConfigSource("worker")

	go func() {
		logger.InfoX("asynq worker started", zap.String("redis", deps.Cfg.Asynq.RedisAddr))
		if err := deps.Server.Run(deps.Mux); err != nil {
			logger.ErrorX("asynq server error", zap.Error(err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	deps.Server.Shutdown()
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
