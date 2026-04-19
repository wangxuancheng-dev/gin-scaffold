# 全局限流

## 行为概述

- 中间件在 `routes/router.go` 注册；若 bootstrap 注入了自定义 `limiter.Backend`（如 Redis），则使用 **`LimiterWithBackend`**。
- 默认算法：**每 IP + 每路由** 两层检查（见 `middleware/limiter.go` 与 `pkg/limiter`）。一次请求需 **先后通过** IP 维度与路由维度，任一不通过即返回 **429**。
- **内存模式（`mode: memory`）**：默认 **`ip_rps` / `ip_burst`（及路由同理）** 为 `golang.org/x/time/rate` 令牌桶；若配置了 **`ip_max_per_window` / `route_max_per_window`**（>0），对应维度改为 **固定窗口计数**（与 Redis 语义对齐，见下节）。
- **Redis 模式（`mode: redis`）**：固定窗口计数，多副本共享（见 `pkg/limiter/redis_store.go`）；亦可对某一维使用 `*_max_per_window` 直接指定每窗口上限。

## 每窗口最大次数（`ip_max_per_window` / `route_max_per_window`）

用于表达「**每 `window_sec` 秒内，该维度最多 N 次**」的**固定窗口**近似（memory / redis 同一套配置键）：

| 配置 | 含义 |
|------|------|
| `window_sec` | 窗口长度（秒）。**redis 必填**；**memory** 在任一大于 0 的 `*_max_per_window` 时也**必填**。 |
| `ip_max_per_window` | >0：每 IP 每窗口最多次数；**0**：IP 维仍用 `ip_rps` / `ip_burst` 令牌桶。 |
| `route_max_per_window` | >0：每路由键每窗口最多次数；**0**：路由维仍用 `route_rps` / `route_burst`。 |

两维可**混合**（例如仅 IP 用窗口、路由用令牌桶）。固定窗口在窗口交界处可能出现「双倍突发」，属算法特性；要更平滑可只用令牌桶或前置网关滑动窗口。

- **memory**：进程内按 key 记录当前窗口 slot 与计数，**重启清零**。
- **redis**：`*_max_per_window` > 0 时该维使用配置值作为硬上限；为 0 时该维仍用 `ceil(rps * window_sec) + burst` 公式。

代码入口：`limiter.NewStoreWithOptions`（`pkg/limiter/store_options.go`）、`bootstrap` 对 `NewRedisStore` 的装配。

## 两层限流分别约束什么

| 维度 | 含义 | 限流键 |
|------|------|--------|
| **IP** | 同一客户端 IP 在单位时间内的总请求量（**该 IP 访问任意接口**都会消耗同一 IP 桶令牌） | `c.ClientIP()`（与 `client_ip` 中间件一致；反代场景需配置 Gin 可信代理，见 [middleware-reference](/guide/middleware-reference)） |
| **Route** | 某一「HTTP 方法 + 路由模板」在全站的总请求量（**所有 IP 共用**该路由的一条计数/桶） | 见下文「路由键」 |

因此：

- `ip_rps` / `ip_burst`：控制**单个用户（IP）**整体刷接口不要太猛。
- `route_rps` / `route_burst`：控制**整条接口**被全站打爆的上限（适合保护下游、写库等），不是「每用户每路由」。

## 路由键长什么样（与 OpenAPI / 排障相关）

中间件里路由维度的键为：

```text
<HTTP 方法> + " " + <Gin 注册的模板路径>
```

- 优先使用 **`c.FullPath()`**（例如 `POST /api/v1/client/orders/:id/submit`）。即 **带参数占位符**，同一路由模板无论 `id` 取何值，共用同一条「路由」限流。
- 若 `FullPath()` 为空（极少数未走 Gin 路由表的情况），会退回 **`c.Request.URL.Path`**（实际请求路径）。

你在日志或 429 排查时，可以把「当前模板路径」理解为上述第二段。

## 自定义限流键（`LimiterKeys`）

本仓库对应能力：

| API | 说明 |
|-----|------|
| `middleware.LimiterWithBackendKeys(b, keys)` | 与 `LimiterWithBackend` 相同，但 `keys` 非 nil 时可改键 |
| `middleware.LimiterWithStoreKeys(store, keys)` | 内存 `Store` + 自定义键的便捷封装 |

`keys` 为 `&middleware.LimiterKeys{ IPKey: ..., RouteKey: ... }`：

- **`IPKey`**：传给 `Backend.AllowIP` 的字符串；为 nil 或返回**仅空白**时，使用 **`c.ClientIP()`**。
- **`RouteKey`**：传给 `Backend.AllowRoute` 的字符串；为 nil 或返回**仅空白**时，使用上文默认的「方法 + 模板路径」。

典型用法（**按登录用户**限流下单；用户标识须来自 **JWT 等可信上下文**，勿直接用未鉴权 query 当唯一键以免被恶意刷桶）：

```go
orders.Use(middleware.LimiterWithBackendKeys(lim, &middleware.LimiterKeys{
    RouteKey: func(c *gin.Context) string {
        if cl, ok := middleware.Claims(c); ok && cl != nil && cl.UserID != 0 {
            return "order:submit:user:" + strconv.FormatInt(cl.UserID, 10)
        }
        return "" // 回退为默认路由键（匿名仍按模板路径维限流）
    },
}))
```

若希望 **整段逻辑只用一个自定义维度**，可让 `IPKey` 返回常量（例如 `"global"`）并把配额全压在 `RouteKey` 上，或反之；仍会通过 `AllowIP` 与 `AllowRoute` 两个桶，需理解两层仍各自消耗令牌。

中间件在 **handler 绑定 JSON 之前**执行，因此 **`RouteKey` 一般不要用未读过的 POST body**（避免与 `ShouldBindJSON` 抢 `Body`）；需要「按请求体某字段」限流时，更稳妥是在 handler 内用 Redis 自管计数，或把关键字段放在 **路径 / Header / JWT** 供 key 使用。

## 示例：要对「订单下单」接口限流

下面分三种常见需求说明；当前 **YAML 不能为单个路径单独配置不同的 `route_rps`**，所有路由共用同一组数字，但 **每个路由模板各自有一个路由桶**（键不同，桶互不干扰）。

### 1）与全站一致：只调配置即可

若 `POST /api/v1/client/orders`（示例路径）与全站共用同一套 `ip_*` / `route_*` 即可：在 `configs/app.yaml`（或环境对应配置）里调整 `limiter`。**该路由**会得到自己的「路由桶」，但 **速率/突发与配置里其它路由相同**。

下单通常写操作多、希望更紧时，在不大改其它接口的前提下，可把全局 `route_rps` / `ip_rps` 略调低，并观察 429 与正常突发。

### 2）只要「下单」比其它接口更严：给该路由组加第二层中间件

全局限流仍会先执行。可在 **仅包含下单相关路由** 的 `gin.RouterGroup` 上再挂一个 `middleware.LimiterWithStore`（或 `LimiterWithBackend`），用 **另一套** `NewStore(ipRPS, ipBurst, routeRPS, routeBurst)`。

注意：第二层默认也会再做 **IP + 路由** 两次检查。若你只想**额外收紧路由**、不想「同一 IP 被扣两次全局限流」，第二层可把 `ip_rps` / `ip_burst` 设得足够大（等价于不对 IP 二次收紧），只把 `route_rps` / `route_burst` 设小。

示例（仅说明写法；路径与 Handler 需按你项目实际注册）：

```go
import (
    "github.com/gin-gonic/gin"

    "gin-scaffold/middleware"
    "gin-scaffold/pkg/limiter"
)

func registerOrderRoutes(clientAuth *gin.RouterGroup /* , h 你的 Handler */) {
    // 仅作用于本 Group 下的路由；在全局 Limiter 之后执行（更严）。
    orders := clientAuth.Group("/orders")
    orders.Use(middleware.LimiterWithStore(limiter.NewStore(
        1_000_000, 1_000_000, // IP：此处放大，避免与全局限流叠加过狠
        3, 8,               // 路由：例如该组下各模板路径各自约 3 rps、burst 8（按业务调）
    )))
    // orders.POST("", h.Create)
    // orders.POST("/:id/submit", h.Submit)
}
```

多实例部署时，第二层也应使用 **与全局一致的 Redis Backend**（从 bootstrap 注入或自行构造 `limiter.NewRedisStore(...)`），否则进程内第二套内存桶无法跨实例。

### 3）与登录防爆破、网关限流的关系

- **全局限流**：保护整个 API 面，防误刷与过载。
- **`platform.login_security`**：仅针对登录失败与锁定（Redis），见 [platform](/guide/platform)。
- 若前置 **API 网关** 已按路径限流，应用内可放宽全局参数，或对敏感路径单独加 `Group` 级中间件（如上）。

## 429 响应

超过限流时返回 HTTP **429**，业务码 **`too_many_req`**，文案为 `too many requests`（IP 触发）或 `route rate limited`（路由触发）。前端可按 429 做退避或提示。

## 调参提示

- `ip_burst` / `route_burst` 过小会导致正常突发请求 429；过大则防护减弱。
- **`route_rps` 是「全站该路由模板」合计**，多用户抢购时容易触顶；若只想「每用户下单频率」，请用 **`LimiterKeys.RouteKey`**（或 `IPKey`）把桶拆成 `user:123` 这类键，见上文「自定义限流键」。

## 配置示例（`configs/*.yaml`）

**单机开发（内存）**

```yaml
limiter:
  mode: memory
  ip_rps: 20
  ip_burst: 40
  route_rps: 50
  route_burst: 80
```

**多副本（Redis 共享计数）**

```yaml
limiter:
  mode: redis
  window_sec: 1
  redis_key_prefix: "app:ratelimit:"
  ip_rps: 50
  ip_burst: 100
  route_rps: 80
  route_burst: 160
```

`window_sec` 在 `redis` 模式下必填；前缀空时会回退 `platform.cache.key_prefix`（见配置校验逻辑）。环境变量映射见 [environment-variables](/guide/environment-variables)。

## 相关代码位置

| 内容 | 路径 |
|------|------|
| Gin 中间件 | `middleware/limiter.go` |
| 内存令牌桶 | `pkg/limiter/limiter.go` |
| Redis 窗口 | `pkg/limiter/redis_store.go` |
| 注册与 Redis 注入 | `routes/router.go`、`internal/app/bootstrap/bootstrap.go` |
