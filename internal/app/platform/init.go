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
	notify.SetDefault(buildDefaultNotifier(cfg))
}

func splitNotifyDriverTokens(driver string) []string {
	raw := strings.TrimSpace(driver)
	if raw == "" {
		return []string{"log"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return []string{"log"}
	}
	return out
}

func buildDefaultNotifier(cfg *config.App) notify.Notifier {
	tokens := splitNotifyDriverTokens(cfg.Platform.Notify.Driver)
	var chain notify.Chain
	for _, t := range tokens {
		switch t {
		case "noop":
			chain = append(chain, notify.Noop{})
		case "log":
			chain = append(chain, notify.LogNotifier{})
		case "smtp":
			chain = append(chain, notify.NewSMTPNotifier(cfg.Platform.Notify.SMTP))
		case "webhook":
			chain = append(chain, notify.NewWebhookNotifier(cfg.Platform.Notify.Webhook))
		}
	}
	if len(chain) == 0 {
		return notify.LogNotifier{}
	}
	if len(chain) == 1 {
		return chain[0]
	}
	return chain
}
