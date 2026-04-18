# 日志与可观测

## 日志能力

- 结构化日志（Zap）
- 全局与单文件轮转策略
- 自定义日志通道
- 支持按时区做按天切割
- 请求链路字段：`request_id` / `trace_id`（响应体见 `api/response` 的 `fillTrace`）

## 配置项速查（`log:`）

| 键 | 说明 |
|----|------|
| `level` | `debug` / `info` / `warn` / `error`，控制应用主日志器 |
| `dir` | 日志根目录（生产常为 `/var/log/app`） |
| `app_file` / `access_file` / `error_file` | 三套输出文件名（相对 `dir`） |
| `rotation_mode` | 全局默认轮转：`size` \| `daily` \| `none` |
| `*_rotation_mode` | 可对 app / access / error 单独覆盖 |
| `max_size_mb` / `max_backups` / `max_age_days` / `compress` | 与 Lumberjack 行为一致 |
| `console` | 是否额外输出到 stdout（生产常为 `false`） |
| `channels` | 具名通道：见下文「自定义通道」 |

## 代码中使用（`pkg/logger`）

初始化在 **`bootstrap.InitServer`**（及 worker 路径）中调用 **`logger.Init(&cfg.Log)`**；退出前 **`logger.Sync()`**。

### 推荐 API

| API | 用途 |
|-----|------|
| `logger.InfoX(msg, zap.Field...)` | 业务 Info（写入 **app** 核心） |
| `logger.WarnX` / `logger.ErrorX` | 告警与错误 |
| `logger.DebugX` | 仅当 `log.level=debug` 时才有意义 |
| `logger.L()` | 需要完整 `*zap.Logger` 时（如第三方库注入） |
| `logger.Channel("daily", "optional.log")` | 写入 `log.channels` 中声明的通道；第二个参数可覆盖文件名 |

### 与 HTTP 请求联动

- **访问日志**：由中间件 `middleware.AccessLog` 写入 **`logger.Access()`**，字段含 `request_id`、`method`、`path`、`status`、`latency`、`client_ip`。
- **业务日志**里建议自行带上 **`request_id`**（从 `gin.Context` 的 `GetString("request_id")` 或统一封装）以便与 access / 审计对齐。
- **链路追踪**：启用 OTel 且中间件生效时，响应 JSON 会带 `trace_id`；日志侧可额外 `zap.String("trace_id", ...)` 与 OTLP 后端关联（按需）。

### 自定义通道示例（`configs/app.yaml` 片段）

```yaml
log:
  channels:
    daily:
      file: task_scheduler.log
      level: info
      rotation_mode: daily
```

代码中：

```go
logger.Channel("daily").Info("sync scheduled tasks ok")
```

任务调度器已使用 `logger.Channel("daily", "task_scheduler.log")` 写调度类日志，见 `internal/job/scheduler`。

## 轮转策略

- `size`: 按大小切割
- `daily`: 按天切割
- `none`: 不切割

可全局配置，也可对 `app/access/error` 单独覆盖。

## 可观测能力

- 健康检查：`/livez`、`/readyz`、`/health`
- 指标：`/metrics`
- 链路追踪：OpenTelemetry（可开关）
- 告警规则模板：`deploy/observability/prometheus-rules.example.yml`
- 看板模板：`deploy/observability/grafana-dashboard-ops.example.json`

## 建议实践

- 生产关闭 `debug`
- 为关键业务操作使用独立 channel（如审计日志）
- 把告警基于 `error` 日志和任务失败日志建立起来
- 在监控中至少覆盖：`5xx 比例`、`P95 延迟`、`队列积压`、`DB 连接池利用率`
- **勿在热路径打印超大结构体**；必要时分字段或采样。
- **敏感信息**（密码、token）禁止写入日志字段。
