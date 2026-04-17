# 平台横切能力（审计 / 幂等 / 缓存 / 事件 / 通知）

与具体业务模块解耦的「框架向」能力，默认关闭或低侵入，可按需在 YAML 中打开。

## 配置入口

顶层 `platform:`，环境变量见 `config/loader.go` 中 `platform.*` 绑定（如 `PLATFORM_AUDIT_ENABLED`）。

## 审计日志（`platform.audit`）

- `enabled: true` 时，对 `POST`/`PUT`/`PATCH`/`DELETE` 异步写入表 `audit_logs`（需执行迁移 `202504171400_create_audit_logs`）。
- 记录：`request_id`、用户（JWT 或 anonymous）、路径、查询串、状态码、耗时、客户端 IP；**不落库请求体**。
- `/livez`、`/readyz`、`/health`、`/swagger`、`/metrics`、`/debug` 等路径自动跳过。

## 幂等（`platform.idempotency`）

- `enabled: true` 且请求带 `X-Idempotency-Key` 时，对 **`Content-Length` 已知的 JSON POST**、路径前缀 `/api/v1/` 的请求做 Redis 缓存。
- 成功（2xx）且 `Content-Type` 为 JSON 的响应体会被缓存，重复请求直接重放；并发相同指纹返回 **409**。
- `multipart`、过大请求体、未知 `Content-Length` 会跳过幂等逻辑。
- 指纹包含：用户维度（JWT `uid` 或 `anon`）、幂等键、路径、**原始请求体**。

## 缓存前缀（`platform.cache`）

- `pkg/cache` 的 `NewFromConfig()` 使用 `key_prefix` 拼接业务 Redis 键（默认 `app:`）。

## 事件总线（`pkg/eventbus`）

- 进程内**同步**派发；耗时逻辑请投递 Asynq。
- 在 `internal/app/platform.Init` 中重置默认总线；业务可 `eventbus.Default().On("name", handler)` 订阅。
- 示例：`user.registered` 在用户注册成功后由 handler 发出。

## 通知（`platform.notify` + `pkg/notify`）

- `driver: log`：写入应用日志（默认）。
- `driver: noop`：丢弃。
- 示例：注册成功后 `notify.Default().Notify(...)`。

## 策略辅助（`pkg/policy`）

- `policy.SameUser(actorID, ownerID)`：判断资源是否属于当前用户，可与 RBAC 组合使用。
