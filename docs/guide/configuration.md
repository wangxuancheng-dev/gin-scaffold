# 配置说明

配置来源优先级（高 -> 低）：

1. 运行时环境变量
2. `configs/app.{env}.{profile}.yaml`
3. `configs/app.{env}.yaml`
4. `configs/app.yaml`

## 关键配置组

## `http`

- `host`
- `port`
- `read_timeout_sec`
- `read_header_timeout_sec`
- `write_timeout_sec`
- `idle_timeout_sec`
- `shutdown_timeout_sec`

## `db`

- `driver`: `mysql` 或 `postgres`
- `dsn`
- `time_zone`: 可被 `TIME_ZONE` 覆盖

## `log`

- `rotation_mode`: `size | daily | none`
- `app/access/error` 可单独覆盖轮转策略
- `channels` 支持自定义日志通道

## `scheduler`

- `enabled`
- `with_seconds`
- `log_retention_days`
- `lock_enabled`
- `lock_ttl_seconds`
- `lock_prefix`

## 启动校验（Fail Fast）

服务启动时会进行关键配置校验，校验失败直接退出，并聚合输出全部错误项，避免带病运行。
