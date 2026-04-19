# 异步队列（Asynq）

## 进程模型

- **生产者**：HTTP 服务内通过 `internal/job.Client` 投递任务（bootstrap 中注入到需要队列的 service）。
- **消费者**：同一二进制 **`./bin/server worker --env <env>`**（或 `go run ./cmd/server worker`），需与 API **同时长期运行**。

## 配置（`configs/*.yaml` 中 `asynq:`）

| 字段 | 含义 |
|------|------|
| `redis_addr` / `redis_password` / `redis_db` | 与业务 Redis 分离 DB 可避免键冲突 |
| `concurrency` | Worker 并发协程数 |
| `queues` | 多队列权重，如 `critical:8, default:3, low:1` |
| `max_retry` | 单任务最大重试次数 |
| `timeout_sec` | **单个 Asynq 任务**执行超时（与 DB 定时任务 `timeout_sec` 不同） |
| `dedup_window_sec` | 短时间入队去重窗口 |

## 客户端 API（`internal/job`）

- `EnqueueTask` / `EnqueueTaskInQueue`：普通入队；`uniqueWindowSec > 0` 时启用 Asynq `Unique` 去重（短窗口防重复点击）。
- `EnqueueWelcome`：示例任务。
- **延时 / 定时入队**（底层为 Asynq `ProcessIn` / `ProcessAt`，任务先进入 **scheduled** 状态，到点再变为 pending 被 worker 消费）：
  - `EnqueueTaskAfter` / `EnqueueTaskInQueueAfter`：相对当前时间延迟 `delay`（`delay <= 0` 时与立即入队相同）。
  - `EnqueueTaskAt` / `EnqueueTaskInQueueAt`：在绝对时间 `at` 执行（`at` 为零值时与立即入队相同）。
  - `EnqueueUniqueAfter` / `EnqueueUniqueInQueueAfter`、`EnqueueUniqueAt` / `EnqueueUniqueInQueueAt`：在配置的 `asynq.dedup_window_sec` 去重窗口上叠加延迟或绝对时间，适合「同一订单多次触发关单检查」等场景。
- 典型用途：**订单支付超时**（下单后 `EnqueueTaskInQueueAfter(..., 10*time.Minute)`，handler 内幂等判断订单状态再关单）、延迟通知、定时生效活动。
- 注意：`timeout_sec` 仍是**任务开始执行后**的单次执行超时，与「何时开始执行」无关；长延迟任务请确认 Redis 持久化与运维策略。
- 新增任务类型：在 `internal/job` 定义 `TypeXxx` 常量、payload 结构，在 worker `Mux` 注册 handler（见 `internal/app/bootstrap` worker 初始化段）。

## 失败与运维

- 管理端 API：`/api/v1/admin/task-queues/*`（需权限），见 [定时任务与队列文档](/guide/scheduler) 中「死信任务运维」。
- CLI：`go run ./cmd/artisan queue:failed ...`（见 [命令系统](/guide/commands)）。

## 与 DB 定时任务的区别

| | Asynq | DB 定时任务（scheduler） |
|--|--------|---------------------------|
| 触发 | 入队后由 worker 消费；支持 **立即**、**延迟**（`ProcessIn`）、**定时**（`ProcessAt`） | Cron 表达式到点执行 shell/artisan |
| 典型用途 | 异步导出、邮件、慢 IO、**每单一次的延迟关单** | 周期清理、报表、调用 `migrate` 等 |
