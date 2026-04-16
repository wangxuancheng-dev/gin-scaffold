# 定时任务中心

任务中心为数据库驱动，支持后台 CRUD 与即时生效。

## 能力清单

- 任务定义入库：`name/spec/command/enabled`
- 支持执行 shell 命令与 `artisan` 命令
- 支持 `timeout_sec`（`0` 表示不超时）
- 支持并发策略：`forbid | allow`
- 支持手动触发 `run now`
- 执行日志入库，可分页查询
- 支持日志自动清理（保留天数）
- Asynq 任务支持重试上限与归档（死信）处理，可通过 artisan 运维命令查看与重试
- Asynq 支持“短时间去重窗口”（`asynq.dedup_window_sec`），可防止重复点击导致同任务多次入队

## 多实例安全执行

- 本机互斥：防止同实例重复执行
- Redis 分布式锁：防止多实例重复执行
- 锁续期：长任务运行时自动续租

## 典型任务示例

- 数据清理：`artisan data:cleanup`
- 报表汇总：`artisan report:daily --date=2026-04-16`
- 迁移脚本：`./bin/migrate up --env prod --driver mysql --dsn "$DB_DSN"`

## 死信任务运维

```bash
# 查看归档任务（dead letter）
go run ./cmd/artisan queue:failed list --env prod

# 重试某条任务
go run ./cmd/artisan queue:failed retry <task_id> --env prod

# 批量重试前 N 条（默认 20）
go run ./cmd/artisan queue:failed retry-all 50 --env prod
```

## 队列去重建议（防重复点击）

- 资金类任务建议使用 `EnqueueUnique(...)` 入队
- 去重判断基于：`taskType + payload`，因此 payload 里应包含稳定业务键（如 `order_id` 或 `request_id`）
- 去重窗口由 `asynq.dedup_window_sec` 控制（例如 30 秒）

## 配置建议

- `lock_enabled=true`（生产建议开启）
- `lock_ttl_seconds` 根据任务最长耗时设置（并预留续期窗口）
- `log_retention_days` 根据审计周期设置
- `asynq.queues` 支持多队列优先级权重，例如 `critical:8, default:3, low:1`
