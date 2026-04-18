# 全局限流

## 行为概述

- 中间件在 `routes/router.go` 注册；若 bootstrap 注入了自定义 `limiter.Backend`（如 Redis），则使用 **`LimiterWithBackend`**。
- 默认算法：**每 IP + 每路由** 令牌桶（见 `middleware/limiter.go` 与 `pkg/limiter`）。

## 配置 `limiter`

| 场景 | 建议 |
|------|------|
| 本地开发 | `mode: memory` 即可 |
| 多副本生产 | **`mode: redis`**，并设置 `window_sec` 与 `redis_key_prefix`（生产模板已示例） |

## 与登录防爆破的区别

- **全局限流**：保护整个 API 面，防 DDoS / 误刷。
- **`platform.login_security`**：针对登录失败次数与锁定（Redis），见 [platform](/guide/platform)。

## 调参提示

- `ip_burst` / `route_burst` 过小会导致正常突发请求 429；过大则防护减弱。
- 若前置 **API 网关** 已做限流，应用层可酌情放宽或仅保留登录等敏感路径的专项限流（需自行拆分路由组）。
