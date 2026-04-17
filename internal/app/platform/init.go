// Package platform 聚合进程级横切组件初始化（事件总线、通知等）。
package platform

import (
	"strings"

	"gin-scaffold/config"
	"gin-scaffold/pkg/eventbus"
	"gin-scaffold/pkg/notify"
)

// Init 根据配置初始化默认 eventbus、notify（幂等与审计由中间件读 config.Get()）。
func Init(cfg *config.App) {
	eventbus.SetDefault(eventbus.New())
	if cfg == nil {
		notify.SetDefault(notify.LogNotifier{})
		return
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Platform.Notify.Driver)) {
	case "noop":
		notify.SetDefault(notify.Noop{})
	default:
		notify.SetDefault(notify.LogNotifier{})
	}
}
