package limiter

import (
	"context"
	"fmt"
	"math"
	"time"

	appredis "gin-scaffold/pkg/redis"
)

// RedisStore 基于 Redis 固定窗口的近似全局限流（多实例共享）。
type RedisStore struct {
	prefix     string
	windowSec  int
	ipRPS      float64
	ipBurst    int
	routeRPS   float64
	routeBurst int
}

// NewRedisStore 创建 Redis 限流器；windowSec 为计数窗口秒数（建议 1~5）。
func NewRedisStore(prefix string, windowSec int, ipRPS float64, ipBurst int, routeRPS float64, routeBurst int) *RedisStore {
	if windowSec <= 0 {
		windowSec = 1
	}
	if prefix == "" {
		prefix = "app:rl:"
	}
	return &RedisStore{
		prefix:     prefix,
		windowSec:  windowSec,
		ipRPS:      ipRPS,
		ipBurst:    ipBurst,
		routeRPS:   routeRPS,
		routeBurst: routeBurst,
	}
}

func (s *RedisStore) ipLimit() int {
	v := int(math.Ceil(s.ipRPS*float64(s.windowSec))) + s.ipBurst
	if v < 1 {
		v = 1
	}
	return v
}

func (s *RedisStore) routeLimit() int {
	v := int(math.Ceil(s.routeRPS*float64(s.windowSec))) + s.routeBurst
	if v < 1 {
		v = 1
	}
	return v
}

func (s *RedisStore) slot() int64 {
	if s.windowSec <= 0 {
		return time.Now().Unix()
	}
	return time.Now().Unix() / int64(s.windowSec)
}

func (s *RedisStore) allow(ctx context.Context, kind, key string, limit int) bool {
	slot := s.slot()
	redisKey := fmt.Sprintf("%sv1:%s:%d:%s", s.prefix, kind, slot, key)
	n, err := appredis.Incr(ctx, redisKey)
	if err != nil {
		return true
	}
	if n == 1 {
		ttl := time.Duration(s.windowSec*2) * time.Second
		_ = appredis.Expire(ctx, redisKey, ttl)
	}
	return int(n) <= limit
}

// AllowIP 按客户端 IP 维度限流。
func (s *RedisStore) AllowIP(ip string) bool {
	if ip == "" {
		ip = "unknown"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	return s.allow(ctx, "ip", ip, s.ipLimit())
}

// AllowRoute 按路由（方法+路径）维度限流。
func (s *RedisStore) AllowRoute(routeKey string) bool {
	if routeKey == "" {
		routeKey = "unknown"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	return s.allow(ctx, "route", routeKey, s.routeLimit())
}

var _ Backend = (*RedisStore)(nil)
