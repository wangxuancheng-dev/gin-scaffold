// Package limiter 提供基于令牌桶的 IP 与路由级限流封装。
package limiter

import (
	"sync"

	"golang.org/x/time/rate"
)

// Store 保存每个 key 对应的限流器（令牌桶与/或固定窗口计数）。
type Store struct {
	mu        sync.Mutex
	windowSec int

	ipMaxWin int
	ipRPS    float64
	ipBurst  int
	ipTB     map[string]*rate.Limiter
	ipWindow map[string]*slotCount

	routeMaxWin int
	routeRPS    float64
	routeBurst  int
	routeTB     map[string]*rate.Limiter
	routeWindow map[string]*slotCount
}

// NewStore 创建纯令牌桶限流器（与历史行为一致）。
func NewStore(ipRPS float64, ipBurst int, routeRPS float64, routeBurst int) *Store {
	return NewStoreWithOptions(StoreOptions{
		IPRPS: ipRPS, IPBurst: ipBurst, RouteRPS: routeRPS, RouteBurst: routeBurst,
	})
}

// NewStoreWithOptions 创建内存限流器；可对 IP / 路由分别选用令牌桶或固定窗口近似。
func NewStoreWithOptions(opt StoreOptions) *Store {
	s := &Store{
		windowSec:   opt.WindowSec,
		ipMaxWin:    opt.IPMaxPerWindow,
		ipRPS:       opt.IPRPS,
		ipBurst:     opt.IPBurst,
		ipTB:        make(map[string]*rate.Limiter),
		routeMaxWin: opt.RouteMaxPerWindow,
		routeRPS:    opt.RouteRPS,
		routeBurst:  opt.RouteBurst,
		routeTB:     make(map[string]*rate.Limiter),
	}
	if opt.IPMaxPerWindow > 0 {
		s.ipWindow = make(map[string]*slotCount)
	}
	if opt.RouteMaxPerWindow > 0 {
		s.routeWindow = make(map[string]*slotCount)
	}
	return s
}

// AllowIP 按客户端 IP 维度限流。
func (s *Store) AllowIP(ip string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ipMaxWin > 0 {
		if s.ipWindow == nil {
			s.ipWindow = make(map[string]*slotCount)
		}
		return s.allowWindowLocked(s.ipWindow, ip, s.ipMaxWin)
	}
	lim, ok := s.ipTB[ip]
	if !ok {
		lim = rate.NewLimiter(rate.Limit(s.ipRPS), s.ipBurst)
		s.ipTB[ip] = lim
	}
	return lim.Allow()
}

// AllowRoute 按路由（方法+路径）维度限流。
func (s *Store) AllowRoute(routeKey string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.routeMaxWin > 0 {
		if s.routeWindow == nil {
			s.routeWindow = make(map[string]*slotCount)
		}
		return s.allowWindowLocked(s.routeWindow, routeKey, s.routeMaxWin)
	}
	lim, ok := s.routeTB[routeKey]
	if !ok {
		lim = rate.NewLimiter(rate.Limit(s.routeRPS), s.routeBurst)
		s.routeTB[routeKey] = lim
	}
	return lim.Allow()
}

var _ Backend = (*Store)(nil)
