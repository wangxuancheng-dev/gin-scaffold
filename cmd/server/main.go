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
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"gin-scaffold/internal/app/bootstrap"
	"gin-scaffold/internal/config"
	"gin-scaffold/pkg/logger"

	_ "gin-scaffold/docs"
)

func main() {
	var env string
	var profile string
	rootCmd := &cobra.Command{
		Use:   "gin-scaffold",
		Short: "HTTP API 或 Asynq Worker",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	rootCmd.PersistentFlags().StringVar(&env, "env", "dev", "配置环境: dev|test|prod")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "配置画像: 多实例标识，如 order/crm")
	rootCmd.AddCommand(&cobra.Command{
		Use:   "server",
		Short: "启动 HTTP 服务",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(env, profile)
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "worker",
		Short: "启动 Asynq 任务消费者",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorker(env, profile)
		},
	})
	if err := rootCmd.Execute(); err != nil {
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
	printServerStartupSummary(deps.Cfg, addr, len(deps.Engine.Routes()))
	srv := &http.Server{
		Addr:              addr,
		Handler:           deps.Engine,
		ReadTimeout:       time.Duration(deps.Cfg.HTTP.ReadTimeout) * time.Second,
		ReadHeaderTimeout: time.Duration(deps.Cfg.HTTP.ReadHeaderTimeout) * time.Second,
		WriteTimeout:      time.Duration(deps.Cfg.HTTP.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(deps.Cfg.HTTP.IdleTimeout) * time.Second,
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

	shutdownTimeout := deps.Cfg.HTTP.ShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = 10
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(shutdownTimeout)*time.Second)
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

func printServerStartupSummary(cfg *config.App, addr string, routeCount int) {
	if cfg == nil {
		return
	}
	mode := "release"
	if cfg.Debug {
		mode = "debug"
	}
	line := strings.Repeat("=", 72)
	fmt.Printf("\x1b[35m%s\x1b[0m\n", line)
	fmt.Printf("\x1b[35m%-72s\x1b[0m\n", " Gin Scaffold Startup ")
	fmt.Printf("\x1b[35m%s\x1b[0m\n", line)
	fmt.Printf("  \x1b[36mMode\x1b[0m       : %s\n", mode)
	fmt.Printf("  \x1b[36mEnv\x1b[0m        : %s\n", cfg.Env)
	fmt.Printf("  \x1b[36mApp\x1b[0m        : %s\n", cfg.Name)
	fmt.Printf("  \x1b[36mAddress\x1b[0m    : %s\n", addr)
	fmt.Printf("  \x1b[36mRoutes\x1b[0m     : %d\n", routeCount)
	fmt.Printf("  \x1b[36mHTTP Timeout\x1b[0m: read=%ds, header=%ds, write=%ds, idle=%ds, shutdown=%ds\n",
		cfg.HTTP.ReadTimeout,
		cfg.HTTP.ReadHeaderTimeout,
		cfg.HTTP.WriteTimeout,
		cfg.HTTP.IdleTimeout,
		cfg.HTTP.ShutdownTimeout,
	)
	fmt.Printf("\x1b[35m%s\x1b[0m\n", line)
}
