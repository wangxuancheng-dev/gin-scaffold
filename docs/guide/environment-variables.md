# 环境变量与配置覆盖

Viper 将 YAML 键中的 **`.` 转为 `_` 再转大写** 作为环境变量名（见 `config/loader.go` 中 `SetEnvKeyReplacer`）。例如 `http.port` → **`HTTP_PORT`**。

以下列表与 **`bindEnvKeys`** 中显式 `BindEnv` 的键一致；设置同名环境变量可覆盖 `configs/*.yaml`（优先级高于文件）。

## 已绑定环境变量一览

| YAML 键 | 环境变量示例 |
|---------|----------------|
| `env` | `ENV` |
| `name` | `NAME` |
| `debug` | `DEBUG` |
| `http.host` | `HTTP_HOST` |
| `http.port` | `HTTP_PORT` |
| `http.read_timeout_sec` | `HTTP_READ_TIMEOUT_SEC` |
| `http.read_header_timeout_sec` | `HTTP_READ_HEADER_TIMEOUT_SEC` |
| `http.write_timeout_sec` | `HTTP_WRITE_TIMEOUT_SEC` |
| `http.idle_timeout_sec` | `HTTP_IDLE_TIMEOUT_SEC` |
| `http.shutdown_timeout_sec` | `HTTP_SHUTDOWN_TIMEOUT_SEC` |
| `http.max_body_bytes` | `HTTP_MAX_BODY_BYTES` |
| `http.swagger_enabled` | `HTTP_SWAGGER_ENABLED` |
| `log.level` | `LOG_LEVEL` |
| `log.dir` | `LOG_DIR` |
| `log.app_file` | `LOG_APP_FILE` |
| `log.access_file` | `LOG_ACCESS_FILE` |
| `log.error_file` | `LOG_ERROR_FILE` |
| `log.rotation_mode` | `LOG_ROTATION_MODE` |
| `log.app_rotation_mode` | `LOG_APP_ROTATION_MODE` |
| `log.access_rotation_mode` | `LOG_ACCESS_ROTATION_MODE` |
| `log.error_rotation_mode` | `LOG_ERROR_ROTATION_MODE` |
| `log.max_size_mb` | `LOG_MAX_SIZE_MB` |
| `log.max_backups` | `LOG_MAX_BACKUPS` |
| `log.max_age_days` | `LOG_MAX_AGE_DAYS` |
| `log.compress` | `LOG_COMPRESS` |
| `log.console` | `LOG_CONSOLE` |
| `db.driver` | `DB_DRIVER` |
| `db.dsn` | `DB_DSN` |
| `db.max_open_conns` | `DB_MAX_OPEN_CONNS` |
| `db.max_idle_conns` | `DB_MAX_IDLE_CONNS` |
| `db.conn_max_lifetime_sec` | `DB_CONN_MAX_LIFETIME_SEC` |
| `db.conn_max_idle_time_sec` | `DB_CONN_MAX_IDLE_TIME_SEC` |
| `db.slow_threshold_ms` | `DB_SLOW_THRESHOLD_MS` |
| `db.log_level` | `DB_LOG_LEVEL` |
| `redis.addr` | `REDIS_ADDR` |
| `redis.password` | `REDIS_PASSWORD` |
| `redis.db` | `REDIS_DB` |
| `redis.pool_size` | `REDIS_POOL_SIZE` |
| `redis.min_idle_conns` | `REDIS_MIN_IDLE_CONNS` |
| `asynq.redis_addr` | `ASYNQ_REDIS_ADDR` |
| `asynq.redis_password` | `ASYNQ_REDIS_PASSWORD` |
| `asynq.redis_db` | `ASYNQ_REDIS_DB` |
| `asynq.concurrency` | `ASYNQ_CONCURRENCY` |
| `asynq.strict_priority` | `ASYNQ_STRICT_PRIORITY` |
| `asynq.queue` | `ASYNQ_QUEUE` |
| `asynq.max_retry` | `ASYNQ_MAX_RETRY` |
| `asynq.timeout_sec` | `ASYNQ_TIMEOUT_SEC` |
| `asynq.dedup_window_sec` | `ASYNQ_DEDUP_WINDOW_SEC` |
| `jwt.secret` | `JWT_SECRET` |
| `jwt.access_expire_min` | `JWT_ACCESS_EXPIRE_MIN` |
| `jwt.refresh_expire_min` | `JWT_REFRESH_EXPIRE_MIN` |
| `jwt.issuer` | `JWT_ISSUER` |
| `metrics.enabled` | `METRICS_ENABLED` |
| `metrics.path` | `METRICS_PATH` |
| `trace.enabled` | `TRACE_ENABLED` |
| `trace.endpoint` | `TRACE_ENDPOINT` |
| `trace.service_name` | `TRACE_SERVICE_NAME` |
| `trace.insecure` | `TRACE_INSECURE` |
| `i18n.default_lang` | `I18N_DEFAULT_LANG` |
| `limiter.ip_rps` | `LIMITER_IP_RPS` |
| `limiter.ip_burst` | `LIMITER_IP_BURST` |
| `limiter.route_rps` | `LIMITER_ROUTE_RPS` |
| `limiter.route_burst` | `LIMITER_ROUTE_BURST` |
| `snowflake.node` | `SNOWFLAKE_NODE` |
| `rbac.super_admin_user_id` | `RBAC_SUPER_ADMIN_USER_ID` |
| `scheduler.enabled` | `SCHEDULER_ENABLED` |
| `scheduler.with_seconds` | `SCHEDULER_WITH_SECONDS` |
| `scheduler.log_retention_days` | `SCHEDULER_LOG_RETENTION_DAYS` |
| `scheduler.lock_enabled` | `SCHEDULER_LOCK_ENABLED` |
| `scheduler.lock_ttl_seconds` | `SCHEDULER_LOCK_TTL_SECONDS` |
| `scheduler.lock_prefix` | `SCHEDULER_LOCK_PREFIX` |
| `outbound.*` | `OUTBOUND_TIMEOUT_MS` 等（见下） |
| `storage.*` | `STORAGE_ENABLED`、`STORAGE_DRIVER`… |
| `platform.*` | `PLATFORM_AUDIT_ENABLED` 等 |
| `tenant.*` | `TENANT_ENABLED`、`TENANT_HEADER`、`TENANT_DEFAULT_ID` |
| `outbox.*` | `OUTBOX_ENABLED`、`OUTBOX_POLL_INTERVAL_SEC`… |

`outbound` 与 `storage` / `platform` / `outbox` 的完整键名按同样规则转换（`.` → `_`，大写）。

## 特殊绑定

| 说明 | 环境变量 |
|------|-----------|
| 数据库会话时区（覆盖 `db.time_zone`） | **`TIME_ZONE`**（如 `UTC`、`Asia/Shanghai`） |
| 非 dev 环境是否加载 `.env*` | **`LOAD_DOTENV_NON_DEV`**（`true` 时允许加载，见 [配置说明](/guide/configuration)） |

## 未在 `bindEnvKeys` 中逐项绑定的配置

下列项通常只在 **YAML** 中维护，或通过 **Viper 的 `AutomaticEnv` 仍可按键名尝试匹配**（若需强制环境变量覆盖，可在 `bindEnvKeys` 中补注册，属代码改动）：

- `limiter.mode`、`limiter.window_sec`、`limiter.redis_key_prefix`
- `metrics.allowed_networks`（列表型，更适合文件）
- `cors.*`、`i18n.bundle_paths`（数组）
- `db.replicas`（副本 DSN 列表）

生产推荐：**敏感 + 易变**用环境变量（DSN、JWT、Redis 密码），**结构型**保留 YAML。

## 另见

- [配置说明](/guide/configuration) · [配置详解](/guide/configuration-advanced)
- 上线自检脚本：`scripts/deploy/check-prod-env.sh`
