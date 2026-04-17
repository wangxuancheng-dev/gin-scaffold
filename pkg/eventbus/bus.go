// Package eventbus 提供进程内同步事件总线（复杂副作用请投递 Asynq）。
package eventbus

import (
	"context"
	"sync"
)

// Event 领域事件。
type Event struct {
	Name    string
	Payload any
}

// Handler 事件处理函数。
type Handler func(ctx context.Context, e Event)

// Bus 同步事件总线。
type Bus struct {
	mu   sync.RWMutex
	subs map[string][]Handler
}

var (
	defaultMu sync.RWMutex
	defaultB  *Bus
)

// New 创建独立总线实例。
func New() *Bus {
	return &Bus{subs: make(map[string][]Handler)}
}

// SetDefault 设置进程级默认总线。
func SetDefault(b *Bus) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if b == nil {
		defaultB = New()
		return
	}
	defaultB = b
}

// Default 返回默认总线（未设置时为新建实例）。
func Default() *Bus {
	defaultMu.RLock()
	b := defaultB
	defaultMu.RUnlock()
	if b != nil {
		return b
	}
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultB == nil {
		defaultB = New()
	}
	return defaultB
}

// On 订阅事件名（可多次注册多个 Handler）。
func (b *Bus) On(name string, h Handler) {
	if b == nil || h == nil || name == "" {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs[name] = append(b.subs[name], h)
}

// Emit 同步派发事件（按注册顺序执行；panic 由调用方或 Recovery 兜底）。
func (b *Bus) Emit(ctx context.Context, e Event) {
	if b == nil || e.Name == "" {
		return
	}
	b.mu.RLock()
	handlers := append([]Handler(nil), b.subs[e.Name]...)
	b.mu.RUnlock()
	for _, h := range handlers {
		if h != nil {
			h(ctx, e)
		}
	}
}
