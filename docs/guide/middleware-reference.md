# 中间件参考

以下顺序以 `routes/router.go` 中 **`r.Use` 注册顺序**为准（自上而下执行）。

## 全局中间件（所有请求）

| 中间件 | 文件 | 作用 |
|--------|------|------|
| `RequestID` | `middleware/request_id.go` | 生成/透传 `X-Request-ID`，写入响应与日志 |
| `ClientIPContext` | `middleware/client_ip.go` | 将客户端 IP 写入上下文，供限流、登录防护等使用 |
| `Tenant` | `middleware/tenant.go` | 多租户：解析头 / JWT / 默认租户，写入上下文 |
| `BodyLimit` | `middleware/body_limit.go` | 限制请求体字节数（`http.max_body_bytes`） |
| `gin.Logger` | 仅 `debug: true` | Gin 风格请求日志 |
| `Recovery` | `middleware/recovery.go` | Panic 恢复 |
| `AccessLog` | `middleware/access_log.go` | 结构化访问日志 |
| `Idempotency` | `middleware/idempotency.go` | 可选 POST 幂等（见 [platform](/guide/platform)） |
| `Audit` | `middleware/audit.go` | 可选写操作审计 |
| `I18n` | `middleware/i18n.go` | 多语言 Bundle |
| `CORS` | `middleware/cors.go` | 跨域 |
| `Limiter` / `LimiterWithBackend` / `LimiterWithBackendKeys` | `middleware/limiter.go` | 全局限流（memory 或 redis）；`LimiterKeys` 可自定义 IP/路由维度的字符串键 |
| `otelgin` | 可选 | `trace.enabled` 时启用 |
| `Metrics` + 白名单 | `middleware/metrics*.go` | Prometheus 指标路径 |

健康检查、`/swagger`（若开启）、`/metrics`（若开启）等在 `Use` 之后以 `GET` 注册。

## 路由级中间件

| 中间件 | 用法 |
|--------|------|
| `JWTAuth` | 解析 `Authorization: Bearer`，写入 `uid`、租户等到上下文 |
| `RequireRoles` | 角色集合，如 `admin` |
| `RequirePermission` | RBAC 权限点，与菜单/种子数据一致 |

权限检查器在 bootstrap 中注入：`middleware.SetPermissionChecker(...)`。

### 路由组挂载示例（源码摘录）

客户端「公开」与「需 JWT」分组（`routes/client_router.go`）：

```go
client := r.Group("/api/v1/client")
{
    client.POST("/auth/login", user.Login)
}
clientAuth := r.Group("/api/v1/client")
clientAuth.Use(middleware.JWTAuth(jwtMgr))
clientAuth.GET("/users/:id", user.Get)
```

管理端在 `routes/adminroutes/register.go` 内对 `admin` 组依次 `Use(middleware.JWTAuth(...))`、`Use(middleware.RequireRoles("admin"))`，再在子路由文件里挂 `RequirePermission`。

## 新增中间件的建议

1. 纯技术横切 → 放 `middleware/`，在 `routes.Build` 里 `r.Use` 插入到合适位置（注意：越靠前越早执行）。
2. 仅某路由组需要 → 在 `RouterGroup.Use` 上挂，避免全局副作用。
3. 与配置强相关 → 从 `config.Get()` 读取，并在 `Validate` 中校验互斥条件。

## 与审计/幂等的关系

- 审计、幂等中间件内部会读取 `platform.*` 配置；关闭时低开销跳过。
- 详见 [平台横切能力](/guide/platform)。
