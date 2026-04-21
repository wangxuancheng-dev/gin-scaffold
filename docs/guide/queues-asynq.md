# 异步队列（Asynq）

## 概述

- 队列基于 **Redis** 与 **[Asynq](https://github.com/hibiken/asynq)**：API 进程负责 **入队**，独立 **`worker` 进程**负责 **消费**。
- 任务类型由 **字符串常量** + **JSON 载荷** 标识；消费逻辑为 **`asynq.Handler`**，在 `InitWorker` 中注册。

## 进程模型

- **生产者**：HTTP 服务内通过 `internal/job.Client` 投递任务（`internal/app/bootstrap` 里注入到需要队列的 `service`）。
- **消费者**：同一二进制 **`go run ./cmd/server worker --env <env>`**（或 `./bin/server worker`），需与 API **同时长期运行**；否则任务会停在 Redis **`pending` / `scheduled`**，业务表现为「无消费」。

## 端到端：新增一种异步任务


下面以假任务 `invoice:notify` 为例，五步与源码目录一一对应；真实任务可对照 **`user:welcome`**（`internal/job/handler/welcome.go`）。

### 1）任务名 + 载荷

在 `internal/job` 包内增加常量与 JSON 可序列化结构体（与 `task_type.go`、`WelcomePayload` 同风格）：

```go
// task_type.go 中增加
const TypeInvoiceNotify = "invoice:notify"

// 可与 WelcomePayload 同文件或新建 invoice_job.go
type InvoiceNotifyPayload struct {
    InvoiceID int64 `json:"invoice_id"`
}
```

### 2）消费者 Handler

新建 `internal/job/handler/invoice_notify.go`：

```go
package handler

import (
    "context"
    "encoding/json"

    "github.com/hibiken/asynq"

    "gin-scaffold/internal/job"
)

type InvoiceNotifyHandler struct{}

func (InvoiceNotifyHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
    var p job.InvoiceNotifyPayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return err
    }
    // TODO: 查库、发通知、写审计；保持幂等（重试时可能重复执行）
    _ = p.InvoiceID
    return nil
}
```

### 3）Worker 注册（必做，否则永远不执行）

`internal/app/bootstrap/bootstrap.go` 的 **`InitWorker`**：

```go
mux.Handle(job.TypeInvoiceNotify, jobhandler.InvoiceNotifyHandler{})
```

### 4）生产者：Service 注入 `*job.Client` 并入队

- 构造函数增加参数 `q *job.Client`（对照 `internal/service/user_service.go`）。
- `InitServer` 里 `NewXxxService(..., q)` 传入。
- 业务方法中：

```go
payload := job.InvoiceNotifyPayload{InvoiceID: id}
if err := s.queue.EnqueueTaskInQueue(ctx, "default", job.TypeInvoiceNotify, payload, 0); err != nil {
    return err
}
```

### 5）本地验证

```bash
# 终端 1
go run ./cmd/server server --env dev

# 终端 2（必须）
go run ./cmd/server worker --env dev
```

触发你的业务接口后，应能在 worker 日志或断点中看到 `ProcessTask` 执行。

---

## 配置（`configs/*.yaml` 中 `asynq:`）

| 字段 | 含义 |
|------|------|
| `redis_addr` / `redis_password` / `redis_db` | 与业务 Redis 分离 DB 可避免键冲突 |
| `concurrency` | Worker 并发协程数 |
| `queues` | 多队列权重，如 `critical:8, default:3, low:1` |
| `max_retry` | 单任务最大重试次数 |
| `timeout_sec` | **单个 Asynq 任务**执行超时（与 DB 定时任务 `timeout_sec` 不同） |
| `dedup_window_sec` | 短时间入队去重窗口（配合 `EnqueueUnique*`） |

## 客户端 API（`internal/job`）

- `EnqueueTask` / `EnqueueTaskInQueue`：普通入队；`uniqueWindowSec > 0` 时启用 Asynq `Unique` 去重（短窗口防重复点击）。
- `EnqueueWelcome`：内置示例任务。
- **延时 / 定时入队**（底层 `ProcessIn` / `ProcessAt`，先入 **scheduled**，到点再被 worker 拉取）：
  - `EnqueueTaskAfter` / `EnqueueTaskInQueueAfter`：相对当前延迟 `delay`（`delay <= 0` 等价立即）。
  - `EnqueueTaskAt` / `EnqueueTaskInQueueAt`：绝对时间 `at`（零值等价立即）。
  - `EnqueueUniqueAfter` / `EnqueueUniqueInQueueAfter`、`EnqueueUniqueAt` / `EnqueueUniqueInQueueAt`：叠加配置里的 `dedup_window_sec` 去重，适合「同一订单多次触发关单检查」。
- 典型用途：**订单支付超时**（`EnqueueTaskInQueueAfter(..., 10*time.Minute)` + handler 内幂等关单）、延迟通知。
- 注意：`timeout_sec` 是**开始执行之后**的单次超时，与「何时开始执行」无关。
- 新增类型：除上文五步外，无需改数据库；任务元数据在 Redis。

## 代码片段（Go）

以下假定已注入 `*job.Client`（变量名 `q`），与 `internal/service/user_service.go` 一致。

**立即入队（默认队列）**

```go
import "gin-scaffold/internal/job"

payload := job.WelcomePayload{UserID: u.ID, Username: u.Username}
if err := q.EnqueueTask(ctx, job.TypeWelcomeEmail, payload, 0); err != nil {
    return err
}
```

**指定队列 + 短窗口去重**

```go
if err := q.EnqueueTaskInQueue(ctx, "low", job.TypeUserExport, exportPayload, 30); err != nil {
    return err
}
```

**延迟执行（例如 10 分钟后）**

```go
import "time"

if err := q.EnqueueTaskInQueueAfter(ctx, "default", job.TypeWelcomeEmail, payload, 0, 10*time.Minute); err != nil {
    return err
}
```

**延迟 + 配置去重窗口**

```go
if err := q.EnqueueUniqueInQueueAfter(ctx, "default", job.TypeWelcomeEmail, payload, 10*time.Minute); err != nil {
    return err
}
```

**绝对时间执行**

```go
at := time.Date(2026, 4, 20, 9, 0, 0, 0, time.Local)
if err := q.EnqueueTaskAt(ctx, job.TypeWelcomeEmail, payload, 0, at); err != nil {
    return err
}
```

**Worker 注册（摘录）**

```go
mux.Handle(job.TypeWelcomeEmail, jobhandler.WelcomeHandler{})
```

## 失败与运维

- 管理端 API：`/api/v1/admin/task-queues/*`（需权限），见 [定时任务与队列](/guide/scheduler) 中「死信任务运维」。
- CLI：`go run ./cmd/artisan queue:failed ...`（见 [命令系统](/guide/commands)）。

## 与 DB 定时任务的区别

| | Asynq | DB 定时任务（scheduler） |
|--|--------|---------------------------|
| 触发 | 入队后由 worker 消费；支持 **立即**、**延迟**、**定时** | Cron 表达式到点执行 shell/artisan |
| 典型用途 | 异步导出、邮件、慢 IO、**每单一次的延迟关单** | 周期清理、报表、调用 `migrate` 等 |

## 最小可复制验证

```bash
# 终端 1：启动 API
go run ./cmd/server server --env dev

# 终端 2：启动 worker（必须）
go run ./cmd/server worker --env dev

# 终端 3：触发一个会入队的接口（示例）
curl -sS "http://127.0.0.1:8080/api/v1/admin/task-queues/summary" \
  -H "Authorization: Bearer <admin-jwt>"
```

验证点：

- worker 日志出现对应任务 `task_type` 的消费记录；
- Redis 中 `pending/scheduled` 不持续堆积（可配合后台死信接口观察）；
- 若使用 `EnqueueUnique*`，同窗口重复触发不会产生重复任务。

## 常见问题与排查

- API 已调用但任务不执行：通常是没启动 `worker`，或 worker 连接到了错误的 Redis/DB。
- 任务大量积压：先看 `asynq.queues` 权重和 `concurrency`，再核对慢任务是否缺少拆分与幂等。
- 重复点击产生多条任务：改用 `EnqueueUnique*`，并确认 `asynq.dedup_window_sec` > 0。
- 任务频繁超时：检查 `asynq.timeout_sec` 是否过小，并排查 handler 内外部 IO（DB/HTTP）耗时。
- 死信持续增长：使用 [定时任务中心](/guide/scheduler) 的死信运维接口/命令重试，并结合 [日志与可观测](/guide/logging-observability) 指标定位失败原因。
