# 常用包与辅助能力

## `pkg/`（库级、可跨项目复用）

| 包 | 用途 |
|----|------|
| `pkg/cache` | Redis JSON 缓存（前缀见配置） |
| `pkg/db` | GORM 初始化、读写分离钩子 |
| `pkg/redis` | Redis 客户端单例 |
| `pkg/logger` | Zap + 切割策略 |
| `pkg/limiter` | 限流后端实现 |
| `pkg/storage` | 本地 / S3 存储抽象 |
| `pkg/notify` | 通知通道（log/smtp/webhook） |
| `pkg/eventbus` | 进程内同步事件 |
| `pkg/httpclient` | 出站 HTTP（超时、重试、熔断） |
| `pkg/tracer` | OTel 初始化 |
| `pkg/settings` | 系统参数读取（短 TTL） |
| `pkg/httpclient` | 出站 HTTP（见 [出站 HTTP 客户端](/guide/outbound-httpclient)） |
| `pkg/eventbus` | 进程内同步事件总线（Outbox 分发等，见 [platform](/guide/platform)） |
| `pkg/notify` | 邮件/Webhook 通知（由 `platform.notify` 驱动） |
| `pkg/policy` | 资源归属等小策略辅助 |

GORM 事务、租户 scope、只读副本等见 **[数据库与 GORM 实践](/guide/database-patterns)**。

## `internal/pkg/`（与本项目域强相关）

| 包 | 用途 |
|----|------|
| `internal/pkg/jwt` | JWT 签发、解析、刷新存储 |
| `internal/pkg/errcode` | 业务错误码与 `BizError` |
| `internal/pkg/validator` | 请求校验封装 |
| `internal/pkg/i18n` | Handler 内 `T()` 翻译 |
| `internal/pkg/tenant` | DAO 层租户 scope |
| `internal/pkg/snowflake` | ID 生成（**多实例 `snowflake.node` 必须唯一**） |
| `internal/pkg/clientip` | IP 解析辅助 |

## Handler 错误出口

- **禁止**在 `api/handler/**`（除 `error_helper.go`）直接调用 `response.FailHTTP` / `FailBiz`；统一用 `api/handler` 包中 `Fail*` 辅助函数（CI 脚本会检查）。

## Artisan 自定义命令

- 实现 `internal/console.Command` 接口并在 `init` 中 `console.Register`（参考 `internal/console/commands/ping.go`）。
- 入口：`go run ./cmd/artisan run <name> --env dev`。
