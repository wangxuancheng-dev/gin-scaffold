package httpclient

import (
	"sync/atomic"
	"time"

	"gin-scaffold/internal/config"
)

var global atomic.Value // *Client

// InitDefault 使用应用配置初始化全局客户端。
func InitDefault(cfg config.OutboundConfig) {
	global.Store(New(Config{
		Timeout:          time.Duration(cfg.TimeoutMS) * time.Millisecond,
		RetryMax:         cfg.RetryMax,
		RetryBackoff:     time.Duration(cfg.RetryBackoffMS) * time.Millisecond,
		CircuitThreshold: uint32(cfg.CircuitThreshold),
		CircuitOpen:      time.Duration(cfg.CircuitOpenSec) * time.Second,
	}))
}

// Default 返回全局治理客户端。
func Default() *Client {
	v := global.Load()
	if v == nil {
		return New(Config{})
	}
	c, _ := v.(*Client)
	return c
}
