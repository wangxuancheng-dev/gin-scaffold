// Package limiter 提供基于令牌桶的 IP 与路由级限流封装。
package limiter

import (
	"sync"

	"golang.org/x/time/rate"
)

// Store 保存每个 key 对应的限流器。
type Store struct {
	mu                  sync.Mutex
	ip                  map[string]*rate.Limiter
	route               map[string]*rate.Limiter
	ipRPS, routeRPS     float64
	ipBurst, routeBurst int
}

// NewStore 创建限流器集合。
func NewStore(ipRPS float64, ipBurst int, routeRPS float64, routeBurst int) *Store {
	return &Store{
		ip:         make(map[string]*rate.Limiter),
		route:      make(map[string]*rate.Limiter),
		ipRPS:      ipRPS,
		ipBurst:    ipBurst,
		routeRPS:   routeRPS,
		routeBurst: routeBurst,
	}
}

// AllowIP 按客户端 IP 维度限流。
func (s *Store) AllowIP(ip string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	lim, ok := s.ip[ip]
	if !ok {
		lim = rate.NewLimiter(rate.Limit(s.ipRPS), s.ipBurst)
		s.ip[ip] = lim
	}
	return lim.Allow()
}

// AllowRoute 按路由（方法+路径）维度限流。
func (s *Store) AllowRoute(routeKey string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	lim, ok := s.route[routeKey]
	if !ok {
		lim = rate.NewLimiter(rate.Limit(s.routeRPS), s.routeBurst)
		s.route[routeKey] = lim
	}
	return lim.Allow()
}

var _ Backend = (*Store)(nil)
