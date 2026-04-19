# 定时任务中心

任务中心为数据库驱动，支持后台 CRUD 与即时生效。

## 能力清单

- 任务定义入库：`name/spec/command/enabled`
- 支持执行 **`artisan` 命令**；可选在 **`scheduler.shell_commands_enabled=true`** 时执行任意 shell（**生产默认关闭**，见 [安全实践](/guide/security-practices)）
- 支持 `timeout_sec`（`0` 表示不超时）
- 支持并发策略：`forbid | allow`
- 支持手动触发 `run now`
- 执行日志入库，可分页查询
- 支持日志自动清理（保留天数）
- Asynq 任务支持重试上限与归档（死信）处理，可通过 artisan 运维命令查看与重试
- Asynq 支持“短时间去重窗口”（`asynq.dedup_window_sec`），可防止重复点击导致同任务多次入队

## 表达式 `spec` 说明（Cron）

任务调度使用 **[robfig/cron/v3](https://github.com/robfig/cron)**，与配置项 **`scheduler.with_seconds`** 绑定：

| `with_seconds` | 格式 | 字段顺序（从左到右） |
| --- | --- | --- |
| **`false`（默认）** | **5 段** | **分 时 日 月 周** |
| **`true`** | **6 段** | **秒 分 时 日 月 周** |

- **日** = 月内几号（day-of-month），**周** = 星期几（day-of-week）。二者多数场景二选一使用，另一项写 `*`。
- 还支持 **描述符**（如 `@hourly`、`@every 1h`），与 5/6 段表达式可择一使用。
- **解析时区**：与进程 `time.Local` 一致（启动时会按数据库 `time_zone` 等对齐全局时区），写入 `spec` 时请按**服务器本地时区**理解「每天几点」。
- **生效延迟**：调度器约 **每 10 秒** 从数据库同步一次任务；新建/修改后一般不会超过该间隔开始按 `spec` 触发。

### 默认（`with_seconds: false`）常用示例

以下为 **5 段**：`分 时 日 月 周`。

| 需求 | `spec` 示例 | 说明 |
| --- | --- | --- |
| 每分钟一次 | `* * * * *` | 每分钟一次（按「分」字段粒度，与整点对齐方式见 robfig 文档） |
| 每 N 分钟 | `*/5 * * * *` | 每 5 分钟 |
| 每小时整点 | `0 * * * *` | 每小时 0 分 |
| 每 N 小时 | `0 */3 * * *` | 从 0 点起每 3 小时一次（0、3、6…点整） |
| 每天固定时刻 | `30 14 * * *` | 每天 **14:30**（14 点 30 分） |
| 工作日 9 点 | `0 9 * * 1-5` | 周一～五 09:00 |
| 每周一 9 点 | `0 9 * * 1` | 周一 09:00（周字段：`0`=周日，`1`=周一 … `6`=周六） |
| 每月 1 号 0 点 | `0 0 1 * *` | 每月 1 日 00:00 |
| 每月 15 号 10:30 | `30 10 15 * *` | 每月 15 日 10:30 |

**周与日的组合**：若同时约束「几号」和「星期几」，语义为 **OR**（满足其一即运行），容易不符合直觉；通常只约束一项，另一项用 `*`。

### 预置描述符（与 `with_seconds` 无关，可直接用）

| `spec` | 含义（等价 5 段语义） |
| --- | --- |
| `@yearly` / `@annually` | 每年 1 月 1 日 00:00 |
| `@monthly` | 每月 1 日 00:00 |
| `@weekly` | 每周日 00:00 |
| `@daily` / `@midnight` | 每天 00:00 |
| `@hourly` | 每小时 0 分（整点起算） |

### 固定间隔：`@every`

使用 Go **`time.ParseDuration`** 能解析的字符串，例如：

| `spec` | 含义 |
| --- | --- |
| `@every 1m` | 每 1 分钟 |
| `@every 1h30m` | 每 1 小时 30 分钟 |
| `@every 45s` | 每 45 秒（需本库支持该 duration；若解析失败保存任务时会校验失败） |

适合「每隔一段固定 wall 间隔」而不是「对齐到日历整点」的场景。

### 含秒（`with_seconds: true`）—— 6 段表达式

在 **`configs/*.yaml`** 中设置 `scheduler.with_seconds: true` 并重启后，`spec` 必须为 **6 段**：**秒 分 时 日 月 周**。

| 需求 | `spec` 示例 |
| --- | --- |
| 每秒 | `* * * * * *` | 每秒触发（负载极高，慎用） |
| 每 N 秒 | `*/10 * * * * *` | 每 10 秒 |
| 每分钟第 0 秒 | `0 * * * * *` | 每分钟一次，且对齐到秒 0 |
| 每天 09:30:00 | `0 30 9 * * *` | 与 5 段 `30 9 * * *` 对应，增加秒位 |

未开启 `with_seconds` 时 **不要** 写 6 段，否则 API 校验会报 **Cron 表达式非法**。

### 校验与排错

- 后台创建/更新任务时，服务端会用与运行时相同的规则解析 `spec`，非法则拒绝保存。
- 更多语法（如 `L`、`W`、`#`、名称 `MON` 等）以 **robfig/cron** 文档为准：[Cron Expression Format](https://pkg.go.dev/github.com/robfig/cron/v3#hdr-CRON_Expression_Format)。

## 多实例安全执行

- 本机互斥：防止同实例重复执行
- Redis 分布式锁：防止多实例重复执行
- 锁续期：长任务运行时自动续租

## 典型任务示例

- 数据清理：`artisan data:cleanup`
- 报表汇总：`artisan report:daily --date=2026-04-16`
- 迁移脚本：`./bin/migrate up --env prod --driver mysql --dsn "$DB_DSN"`

## 死信任务运维

后台 API（需 admin + 任务权限）：

- `GET /api/v1/admin/task-queues/summary`
- `GET /api/v1/admin/task-queues/failed?queue=default&state=retry`
- `POST /api/v1/admin/task-queues/{queue}/failed/{task_id}/retry`
- `POST /api/v1/admin/task-queues/{queue}/failed/{task_id}/archive`

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

## 多队列优先级使用方式（critical/default/low）

- 在配置中用 `asynq.queues` 设置权重，例如：`critical:8, default:3, low:1`
- 业务入队时必须把任务投到对应队列（例如资金类投 `critical`），否则优先级权重无法生效
- 入队方式：
  - `EnqueueTaskInQueue(ctx, "critical", taskType, payload, uniqueWindowSec)`
  - `EnqueueUniqueInQueue(ctx, "critical", taskType, payload)`
  - **相对/绝对延时入队**（如 10 分钟后关单）：见 [异步队列（Asynq）](/guide/queues-asynq) 中「延时 / 定时入队」；与 DB Cron 定时任务互补。

## 配置建议

- **`shell_commands_enabled`**：生产 **`false`**，任务 `command` 仅写 `artisan <name> [args]`；本地/特殊运维需要跑脚本时再打开。
- `lock_enabled=true`（生产建议开启）
- `lock_ttl_seconds` 根据任务最长耗时设置（并预留续期窗口）
- `log_retention_days` 根据审计周期设置
- `asynq.queues` 支持多队列优先级权重，例如 `critical:8, default:3, low:1`
