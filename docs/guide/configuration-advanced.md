# 配置详解（全量键速查）

与 [配置说明（关键组）](/guide/configuration) 互补：下列条目对应 `config/config.go` 中结构体，YAML 路径与 `config/loader.go` 中 `BindEnv` 一致处可用环境变量覆盖。

## `env` / `name` / `debug`

- `env`：逻辑环境，影响加载 `app.{env}.yaml`。
- `debug`：Gin 模式、部分调试路由、更详细日志行为。

## `jwt`

| 键 | 说明 |
|----|------|
| `secret` | 签名密钥（生产必须强随机，环境变量 `JWT_SECRET`） |
| `access_expire_min` / `refresh_expire_min` | 访问/刷新令牌分钟数 |
| `issuer` | JWT `iss` |

## `redis`

连接池、超时、`db` 索引；与 Asynq 使用 **不同 db** 为常见实践。

## `asynq`

见 [异步队列](/guide/queues-asynq)。

## `cors`

- `allow_origins`：生产勿用 `*` 携带 `allow_credentials`。
- `allow_headers`：需包含租户头（如 `X-Tenant-ID`）若启用多租户。

## `limiter`

| 键 | 说明 |
|----|------|
| `mode` | `memory`（单进程）或 `redis`（多副本共享窗口） |
| `window_sec` | **redis 模式必填**，固定窗口长度 |
| `redis_key_prefix` | 限流键前缀；可空则回退缓存前缀 |
| `ip_rps` / `ip_burst` / `route_rps` / `route_burst` | 令牌桶参数 |

## `tenant`

- `enabled` / `header` / `default_id`：解析顺序见 [platform](/guide/platform)。

## `trace`

- `enabled` / `endpoint`（OTLP HTTP）/ `service_name` / `insecure`。

## `metrics`

- `enabled` / `path` / `allowed_networks`：见 [配置说明](/guide/configuration) 中 metrics 小节。

## `snowflake`

- `node`：0–1023，多实例部署 **必须唯一**，否则 ID 冲突。

## `rbac`

- `super_admin_user_id`：该用户 ID 拥有全部权限且受保护逻辑（见种子与中间件）。

## `outbox`

- `enabled` / `poll_interval_sec` / `batch_size` / `max_attempts` / `retry_backoff_sec` / `publish_mode`（`eventbus` | `http`）等，见 [platform](/guide/platform)。

## `platform.*`

审计、幂等、通知、登录防爆破等，统一见 [平台横切能力](/guide/platform)。

## 启动校验

所有键在 `config.Validate()` 中聚合校验；错误会阻止进程启动（fail-fast）。
