# 常用包与辅助能力

## `pkg/`（库级、可跨项目复用）

| 包 | 用途 |
|----|------|
| `pkg/cache` | Redis JSON 缓存（前缀见配置） |
| `pkg/db` | GORM 初始化、读写分离、`session_tz` |
| `pkg/redis` | Redis 客户端、`singleflight` 缓存封装、分布式锁 |
| `pkg/logger` | Zap + 切割策略、访问/错误日志 |
| `pkg/limiter` | 全局限流（`golang.org/x/time/rate`、Redis 后端）；自定义桶键用 `middleware.LimiterWithBackendKeys` |
| `pkg/storage` | 本地 / S3 存储抽象、预签名 |
| `pkg/notify` | 通知通道（log / smtp / webhook，由 `platform.notify` 驱动） |
| `pkg/eventbus` | 进程内事件总线（Outbox 分发等，见 [platform](/guide/platform)） |
| `pkg/httpclient` | 出站 HTTP（超时、重试、熔断，见 [出站 HTTP 客户端](/guide/outbound-httpclient)） |
| `pkg/tracer` | OTel 初始化 |
| `pkg/settings` | 系统参数读取（短 TTL） |
| `pkg/policy` | 资源归属等小策略辅助 |
| `pkg/loginthrottle` | 登录防护节流 |
| `pkg/metrics` / `pkg/swagger` | 指标与 Swagger 辅助 |
| `pkg/strutil` | 分隔串拆分/拼接（`SplitClean` / `JoinClean`）、`*string` 安全取值 |
| `pkg/sliceutil` | 去重保序、过滤、`Coalesce` 默认值 |
| `pkg/numconv` | 从字符串安全解析数字（失败/空串用默认值） |

GORM 事务、租户 scope、只读副本等见 **[数据库与 GORM 实践](/guide/database-patterns)**。

## 小示例（`pkg/sliceutil` / `pkg/numconv`）

```go
import (
    "gin-scaffold/pkg/numconv"
    "gin-scaffold/pkg/sliceutil"
)

tags := sliceutil.UniqueStable([]string{"a", "b", "a"})
_ = numconv.ParseInt64("not-a-number", -1) // => -1
```

## `internal/pkg/`（与本项目域强相关）

| 包 | 用途 |
|----|------|
| `internal/pkg/jwt` | JWT 签发、解析、刷新存储、黑名单 |
| `internal/pkg/errcode` | 业务错误码与 `BizError` |
| `internal/pkg/validator` | 请求校验封装 |
| `internal/pkg/i18n` | Handler 内 `T()` 翻译 |
| `internal/pkg/tenant` | DAO 层租户 scope |
| `internal/pkg/snowflake` | ID 生成（**多实例 `snowflake.node` 必须唯一**） |
| `internal/pkg/clientip` | IP 解析辅助 |
| `internal/pkg/timefmt` | `ParseRFC3339`、`FormatPtr`（`nil` 或零值 → `""`，否则 RFC3339） |
| `internal/pkg/websocket` | WebSocket Hub（见 [SSE/WebSocket](/guide/realtime-sse-websocket)） |

## 助手封装：要不要加、加在哪

- **够用的现状**：按 **领域** 拆在 `pkg/*` 与 `internal/pkg/*`（JWT、租户、错误码、校验、时间解析等），比一个大杂烩 `util` 更可维护。
- **何时新建小包**：同一逻辑在 **≥3 处** 复制、或带 **明确策略**（租户、时区、错误映射）时再抽到 `internal/pkg/xxx`；避免仅为别名而封装（例如不要再包一层「等于 `time.RFC3339`」的常量）。
- **不必做的**：巨型 `helpers`、与业务无关的「全家桶」字符串/切片库；切片/映射优先标准库（见下节）。

### 字符串按符号切 / 切片拼成串，新增放哪？

- **默认**：直接用标准库 **`strings.Split`**、**`strings.Join`**；`[]string` ↔ 分隔串没有项目特殊规则时，**不必**为封装而封装。
- **多处复用、且与业务无关**（例如统一的 trim 空段、去重、转 `[]int`）：在已有 **`pkg/strutil`** / **`pkg/sliceutil`** 上继续加函数（单文件即可），供 `internal/` 与 `api/` 共用。
- **分隔规则是本产品约定**（例如「标签存库必须是 `|`、空项禁止」）：放在 **`internal/pkg/strutil`**（或离模型最近的一层，如 `internal/model` 旁的小包），避免把域规则写进可复用的 `pkg/`。
- **命名**：不要用根上的 `util` / `common`；包名用 **`strutil`**、**`timefmt`** 这种可 grep、可单测的主题名。

## 切片、字符串、时间（风格约定）

- **切片**：排序、包含、去重等用标准库 [`slices`](https://pkg.go.dev/slices)（以及 Go 1.23+ 的 [`maps`](https://pkg.go.dev/maps)）；把 `[]A` 转成 `[]B` 时用 **`make([]B, len(a))` + `for` 预分配**。列表去重保序、过滤、多候选默认值见 **`pkg/sliceutil`**（`UniqueStable` / `Filter` / `Coalesce`）。
- **字符串**：标准库 `strings` / `strconv`；复杂格式化用 `fmt`。标签类「按分隔符切、去空、再拼回」见 **`pkg/strutil`**（`SplitClean` / `JoinClean` / `StringValue`）。
- **时间**：展示与序列化继续用 **`time.RFC3339`** 与 `time.Time.Format`；**解析** HTTP/任务载荷里的 RFC3339 用 **`timefmt.ParseRFC3339`**；可空或未设置时间（`nil` / 零值）写 JSON 可用 **`timefmt.FormatPtr`**。限流等仍用 `golang.org/x/time/rate`（`pkg/limiter`）。
- **数字（query / 表单字符串）**：需要「解析失败就用默认」时用 **`pkg/numconv`**（`ParseInt` / `ParseInt64` / `ParseUint64` / `ParseFloat64`），避免手写重复 `strconv` + `if err`。
- **结构体拷贝**：DTO 层可用 `github.com/jinzhu/copier`（与现有 `internal/api/response` 一致）。

## Handler 错误出口

- **禁止**在 `internal/api/handler/**`（除 `error_helper.go`）直接调用 `response.FailHTTP` / `FailBiz`；统一用 `internal/api/handler` 包中 `Fail*` 辅助函数（CI 脚本会检查）。

## Artisan 自定义命令

- 实现 `internal/console.Command` 接口并在 `init` 中 `console.Register`（参考 `internal/console/commands/ping.go`）。
- 入口：`go run ./cmd/artisan run <name> --env dev`。
