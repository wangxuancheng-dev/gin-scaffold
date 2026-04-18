# 配置说明

配置来源优先级（高 -> 低）：

1. 运行时环境变量
2. `configs/app.{env}.{profile}.yaml`
3. `configs/app.{env}.yaml`
4. `configs/app.yaml`

说明：

- 默认仅 `dev` 环境会自动加载 `.env*` 文件。
- 若希望在 `test/prod` 也读取 `.env.test` / `.env.prod`（例如临时线上压测），可设置环境变量 `LOAD_DOTENV_NON_DEV=true`。
- 即便开启该开关，运行时环境变量依然优先于 `.env*` 与 YAML。

## 关键配置组

## `http`

- `host`
- `port`
- `read_timeout_sec`
- `read_header_timeout_sec`
- `write_timeout_sec`
- `idle_timeout_sec`
- `shutdown_timeout_sec`
- `max_body_bytes`: 请求体大小上限（字节），默认建议 `1048576`（1MB）
- `swagger_enabled`: 是否注册 `/swagger/*`；生产建议 `false`，契约文档可由内网或 CI 产物提供

## `metrics`

- `enabled` / `path`
- `allowed_networks`: CIDR 列表；非空时仅当请求的 **TCP 对端地址**（`RemoteAddr` 解析出的 IP，**不使用** `X-Forwarded-For`）落入任一网段时才返回 Prometheus 指标，否则返回 404。对端为 IPv4-mapped IPv6（`::ffff:x.x.x.x`）时会先规范为 IPv4 再与网段匹配。空列表表示不校验（依赖 Nginx / 网络策略）。生产模板默认填入私网与回环，便于同 VPC 抓取；公网 Prometheus 请置 `[]` 并仅用网关控制访问。

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

## `outbound`

- `timeout_ms`: 下游 HTTP 请求超时
- `retry_max`: 最大重试次数（不含首发）
- `retry_backoff_ms`: 重试退避间隔
- `circuit_threshold`: 连续失败达到阈值后熔断
- `circuit_open_sec`: 熔断打开持续秒数

## `storage`

- `enabled`: 是否启用文件存储
- `driver`: 存储驱动（`local` | `s3` | `minio`，`minio` 与 `s3` 等价）
- `local_dir`: 本地存储根目录
- `sign_secret`: 下载签名密钥（生产必须独立强密钥）
- `max_upload_mb`: 上传大小上限（MB）
- `allowed_ext`: 允许上传扩展名（逗号分隔）
- `allowed_mime`: 允许内容类型（逗号分隔，与 `http.DetectContentType` 嗅探结果比对）
- `url_expire_sec`: 签名下载地址默认过期时间（秒）
- `s3_endpoint` / `s3_region` / `s3_bucket` / `s3_access_key` / `s3_secret_key`：`s3`/`minio` 驱动必填
- `s3_path_style`：是否路径风格访问（MinIO 通常为 `true`）
- `s3_insecure`：是否跳过 TLS 证书校验（仅内网/开发）
- `readyz_check`：`true` 时 `/readyz` 会探测存储（本地目录或 S3 HeadBucket），需 `enabled=true`

## `platform`

审计、幂等、缓存键前缀、通知驱动等；详见 [平台横切能力](/guide/platform)。

- `audit.enabled`：是否将写类 HTTP 请求异步写入 `audit_logs`（需迁移）
- `audit.export_default_days`：审计导出默认时间窗（天）
- `audit.export_max_days`：审计导出最大允许时间窗（天）
- `idempotency.*`：基于 Redis 的 POST 幂等（`X-Idempotency-Key`），见专页说明
- `cache.key_prefix`：`pkg/cache` 使用的 Redis 键前缀
- `notify.driver`：`log`（写应用日志）或 `noop`

## 启动校验（Fail Fast）

服务启动时会进行关键配置校验，校验失败直接退出，并聚合输出全部错误项，避免带病运行。
