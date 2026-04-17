// Package notify 提供可替换的通知抽象（日志、noop；可扩展邮件/IM）。
package notify

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"gin-scaffold/pkg/logger"
)

// Message 一条通知。
type Message struct {
	Channel string            // 业务通道名，如 user、order
	Title   string
	Body    string
	Meta    map[string]string
}

// Notifier 通知发送器。
type Notifier interface {
	Notify(ctx context.Context, msg Message) error
}

var (
	mu       sync.RWMutex
	defaultN Notifier = LogNotifier{}
)

// SetDefault 设置全局 Notifier（进程级，在 bootstrap 中调用）。
func SetDefault(n Notifier) {
	mu.Lock()
	defer mu.Unlock()
	if n == nil {
		defaultN = LogNotifier{}
		return
	}
	defaultN = n
}

// Default 返回当前全局 Notifier。
func Default() Notifier {
	mu.RLock()
	defer mu.RUnlock()
	return defaultN
}

// LogNotifier 将通知写入应用日志（开发/默认）。
type LogNotifier struct{}

// Notify 实现 Notifier。
func (LogNotifier) Notify(ctx context.Context, msg Message) error {
	_ = ctx
	fields := []zap.Field{
		zap.String("channel", msg.Channel),
		zap.String("title", msg.Title),
		zap.String("body", msg.Body),
	}
	for k, v := range msg.Meta {
		fields = append(fields, zap.String("meta_"+k, v))
	}
	logger.L().Info("notify", fields...)
	return nil
}

// Noop 丢弃通知（压测或关闭侧路输出）。
type Noop struct{}

// Notify 实现 Notifier。
func (Noop) Notify(context.Context, Message) error {
	return nil
}
