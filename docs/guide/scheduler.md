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

## 多实例安全执行

- 本机互斥：防止同实例重复执行
- Redis 分布式锁：防止多实例重复执行
- 锁续期：长任务运行时自动续租

## 典型任务示例

- 数据清理：`artisan data:cleanup`
- 报表汇总：`artisan report:daily --date=2026-04-16`
- 迁移脚本：`./bin/migrate up --env prod --driver mysql --dsn "$DB_DSN"`

## 配置建议

- `lock_enabled=true`（生产建议开启）
- `lock_ttl_seconds` 根据任务最长耗时设置（并预留续期窗口）
- `log_retention_days` 根据审计周期设置
