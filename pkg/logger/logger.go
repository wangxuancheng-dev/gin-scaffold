// Package logger 提供基于 Zap + Lumberjack 的结构化日志，支持按级别与用途拆分输出。
package logger

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"gin-scaffold/config"
)

var (
	appL        *zap.Logger
	accessL     *zap.Logger
	errL        *zap.Logger
	channelLogs map[string]*zap.Logger
	once        sync.Once
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

		appCore := zapcore.NewCore(
			jsonEnc,
			writer(filepath.Join(cfg.Dir, cfg.AppFile), cfg, resolveRotationMode(cfg.RotationMode, cfg.AppRotationMode)),
			level,
		)
		errCore := zapcore.NewCore(
			jsonEnc,
			writer(filepath.Join(cfg.Dir, cfg.ErrorFile), cfg, resolveRotationMode(cfg.RotationMode, cfg.ErrorRotationMode)),
			zapcore.ErrorLevel,
		)
		accessCore := zapcore.NewCore(
			jsonEnc,
			writer(filepath.Join(cfg.Dir, cfg.AccessFile), cfg, resolveRotationMode(cfg.RotationMode, cfg.AccessRotationMode)),
			zapcore.InfoLevel,
		)

		cores := []zapcore.Core{appCore, errCore}
		if cfg.Console {
			consoleCore := zapcore.NewCore(jsonEnc, zapcore.AddSync(os.Stdout), level)
			cores = append(cores, consoleCore)
		}
		appL = zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(1))
		accessL = zap.New(accessCore)
		errL = zap.New(zapcore.NewTee(errCore))
		channelLogs = make(map[string]*zap.Logger)
		for name, chCfg := range cfg.Channels {
			file := strings.TrimSpace(chCfg.File)
			if file == "" {
				continue
			}
			chLevel := parseLevelOrDefault(chCfg.Level, level)
			merged := mergeChannelConfig(cfg, chCfg)
			core := zapcore.NewCore(
				jsonEnc,
				writer(filepath.Join(cfg.Dir, file), &merged, resolveRotationMode(merged.RotationMode, "")),
				chLevel,
			)
			channelLogs[name] = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
		}
	})
	return initErr
}

func mergeChannelConfig(base *config.LogConfig, ch config.LogChannelConfig) config.LogConfig {
	merged := *base
	merged.RotationMode = resolveRotationMode(base.RotationMode, ch.RotationMode)
	if ch.MaxSizeMB > 0 {
		merged.MaxSizeMB = ch.MaxSizeMB
	}
	if ch.MaxBackups >= 0 {
		merged.MaxBackups = ch.MaxBackups
	}
	if ch.MaxAgeDays > 0 {
		merged.MaxAgeDays = ch.MaxAgeDays
	}
	if ch.Compress != nil {
		merged.Compress = *ch.Compress
	}
	return merged
}

func parseLevelOrDefault(levelStr string, fallback zapcore.Level) zapcore.Level {
	if lv, err := zapcore.ParseLevel(strings.TrimSpace(levelStr)); err == nil {
		return lv
	}
	return fallback
}

func writer(path string, cfg *config.LogConfig, mode string) zapcore.WriteSyncer {
	switch mode {
	case "daily", "date":
		return newDailyRotateWriter(path, cfg.MaxAgeDays, time.Local)
	case "none", "off":
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			// 配置异常时回退为 stdout，避免启动失败。
			return zapcore.AddSync(os.Stdout)
		}
		return zapcore.AddSync(f)
	}
	lj := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}
	return zapcore.AddSync(lj)
}

func resolveRotationMode(globalMode, fileMode string) string {
	if mode := normalizeRotationMode(fileMode); mode != "" {
		return mode
	}
	if mode := normalizeRotationMode(globalMode); mode != "" {
		return mode
	}
	return "size"
}

func normalizeRotationMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "daily", "date":
		return "daily"
	case "none", "off":
		return "none"
	case "size":
		return "size"
	default:
		return ""
	}
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
	for _, l := range channelLogs {
		if l != nil {
			_ = l.Sync()
		}
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

// Channel 返回自定义通道日志器；未配置时回退主日志器。
func Channel(name string) *zap.Logger {
	if l, ok := channelLogs[name]; ok && l != nil {
		return l
	}
	return appL
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
