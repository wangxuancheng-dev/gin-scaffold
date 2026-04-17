# 日志与可观测

## 日志能力

- 结构化日志（Zap）
- 全局与单文件轮转策略
- 自定义日志通道
- 支持按时区做按天切割
- 请求链路字段：`request_id` / `trace_id`

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
