package limiter

// Backend 抽象限流存储（内存、Redis 等）。
type Backend interface {
	AllowIP(ip string) bool
	AllowRoute(routeKey string) bool
}
