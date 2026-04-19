package limiter

// StoreOptions 内存限流器参数：每个维度可独立选择 **令牌桶** 或 **固定窗口近似**（与 Redis 模式语义对齐）。
//
// 当 IPMaxPerWindow（或 RouteMaxPerWindow）> 0 时，该维度使用固定窗口计数：每 WindowSec 秒内最多允许该次数；
// 为 0 时该维度使用 IPRPS/IPBurst（或 RouteRPS/RouteBurst）令牌桶。
// 启用任一 max_per_window 时 WindowSec 必须 > 0（由 config 校验）。
type StoreOptions struct {
	WindowSec int

	IPMaxPerWindow    int
	RouteMaxPerWindow int

	IPRPS, RouteRPS     float64
	IPBurst, RouteBurst int
}
