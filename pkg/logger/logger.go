// Package logger 提供基于 Zap + Lumberjack 的结构化日志，支持按级别与用途拆分输出。
package logger

import (
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"gin-scaffold/config"
)

var (
	appL    *zap.Logger
	accessL *zap.Logger
	errL    *zap.Logger
	once    sync.Once
)

// Init 根据配置初始化全局日志器（应用、访问、错误三套 Writer）。
func Init(cfg *config.LogConfig) error {
	var initErr error
	once.Do(func() {
		if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
			initErr = err
			return
		}
		encCfg := zap.NewProductionEncoderConfig()
		encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		jsonEnc := zapcore.NewJSONEncoder(encCfg)

		level, err := zapcore.ParseLevel(cfg.Level)
		if err != nil {
			level = zapcore.InfoLevel
		}

		appCore := zapcore.NewCore(jsonEnc, writer(filepath.Join(cfg.Dir, cfg.AppFile), cfg), level)
		errCore := zapcore.NewCore(jsonEnc, writer(filepath.Join(cfg.Dir, cfg.ErrorFile), cfg), zapcore.ErrorLevel)
		accessCore := zapcore.NewCore(jsonEnc, writer(filepath.Join(cfg.Dir, cfg.AccessFile), cfg), zapcore.InfoLevel)

		cores := []zapcore.Core{appCore, errCore}
		if cfg.Console {
			consoleCore := zapcore.NewCore(jsonEnc, zapcore.AddSync(os.Stdout), level)
			cores = append(cores, consoleCore)
		}
		appL = zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(1))
		accessL = zap.New(accessCore)
		errL = zap.New(zapcore.NewTee(errCore))
	})
	return initErr
}

func writer(path string, cfg *config.LogConfig) zapcore.WriteSyncer {
	lj := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}
	return zapcore.AddSync(lj)
}

// Sync 刷盘，进程退出前调用。
func Sync() {
	if appL != nil {
		_ = appL.Sync()
	}
	if accessL != nil {
		_ = accessL.Sync()
	}
	if errL != nil {
		_ = errL.Sync()
	}
}

// L 返回应用主日志器。
func L() *zap.Logger {
	return appL
}

// Access 返回访问日志专用日志器。
func Access() *zap.Logger {
	return accessL
}

// ErrorL 返回错误聚合日志器。
func ErrorL() *zap.Logger {
	return errL
}

// InfoX 带 RequestID 等扩展字段的 Info 日志。
func InfoX(msg string, fields ...zap.Field) {
	if appL == nil {
		return
	}
	appL.Info(msg, fields...)
}

// ErrorX 带扩展字段的 Error 日志。
func ErrorX(msg string, fields ...zap.Field) {
	if appL == nil {
		return
	}
	appL.Error(msg, fields...)
}

// WarnX 带扩展字段的 Warn 日志。
func WarnX(msg string, fields ...zap.Field) {
	if appL == nil {
		return
	}
	appL.Warn(msg, fields...)
}

// DebugX 带扩展字段的 Debug 日志。
func DebugX(msg string, fields ...zap.Field) {
	if appL == nil {
		return
	}
	appL.Debug(msg, fields...)
}
