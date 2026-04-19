# 开发手册（总览）

建议阅读顺序：新人 **[新人入门](/guide/onboarding)** → **[快速开始](/guide/quick-start)** → **[目录结构](/guide/directory-structure)** → 按需跳转到下表各章。

## 架构与工程

| 主题 | 说明 |
|------|------|
| [目录结构与分层](/guide/directory-structure) | `cmd/`、`internal/`、`api/`、`routes/`、`pkg/` 职责与依赖方向 |
| [路由与分组](/guide/routing) | `/api/v1/client`、`/api/v1/admin`、JWT 分组、如何加路由 |
| [中间件参考](/guide/middleware-reference) | 全局中间件顺序、租户、审计、幂等、限流等 |
| [错误与响应](/guide/error-handling) | 统一 `Body`、`errcode`、handler 辅助函数 |

## 配置与横切

| 主题 | 说明 |
|------|------|
| [配置说明（关键组）](/guide/configuration) | 加载顺序、常用块速览 |
| [配置详解（全量键）](/guide/configuration-advanced) | `jwt` / `redis` / `asynq` / `cors` / `limiter` / `tenant` / `trace` 等 |
| [环境变量绑定一览](/guide/environment-variables) | `bindEnvKeys` 与 `TIME_ZONE` 等；未绑定项说明 |
| [平台横切能力](/guide/platform) | 审计、幂等、缓存前缀、Outbox、通知、登录防护 |
| [出站 HTTP 客户端](/guide/outbound-httpclient) | 调用下游 API 的超时、重试、熔断（`outbound` 配置） |

## 数据、缓存、队列

| 主题 | 说明 |
|------|------|
| [数据库迁移与填充](/guide/database-and-migrations) | `cmd/migrate`、schema/seed、MySQL / PostgreSQL |
| [数据库与 GORM 实践](/guide/database-patterns) | 事务、租户 scope、副本、慢 SQL |
| [缓存使用](/guide/caching) | `pkg/cache`、键前缀、JSON 存取 |
| [异步队列（Asynq）](/guide/queues-asynq) | Worker、Client、多队列、延时/定时入队、超时与去重 |
| [定时任务中心](/guide/scheduler) | DB 驱动 Cron、手动执行、`timeout_sec`、锁 |

## 命令行与生成

| 主题 | 说明 |
|------|------|
| [命令系统](/guide/commands) | `server` / `migrate` / `artisan` / 与定时任务 `command` 联动 |
| [代码生成（CRUD）](/guide/codegen) | `cmd/gen` 用法与字段 DSL |
| [生成器走读](/guide/codegen-walkthrough) | 从命令到落文件的步骤说明 |

## 文件、实时通道、国际化、限流

| 主题 | 说明 |
|------|------|
| [文件存储](/guide/file-storage) | 本地上传、S3/MinIO、预签名、下载验签 |
| [SSE 与 WebSocket](/guide/realtime-sse-websocket) | 演示端点、**生产化 WebSocket/Hub**、SSE 反代注意 |
| [本地化（i18n）](/guide/i18n) | 语言包、`Accept-Language`、响应文案 |
| [全局限流](/guide/rate-limiting) | memory / redis 模式、与多实例关系 |

## 安全、可观测、测试、部署

| 主题 | 说明 |
|------|------|
| [安全实践](/guide/security-practices) | JWT、RBAC、Swagger/Metrics、密钥与上线清单 |
| [RBAC 与权限](/guide/rbac-and-permissions) | `RequirePermission`、超管、seed、检查器注入 |
| [管理端 API 总览](/guide/admin-api-overview) | 模块与路由文件索引；字段细节以 Swagger 为准 |
| [日志与可观测](/guide/logging-observability) | **Zap 在代码中的用法**、访问日志、轮转、自定义通道、Prometheus、OTel |
| [测试指南](/guide/testing-guide) | 单元测试、集成测试、build tag、CI、本地 `golangci-lint` 脚本 |
| [常用包与辅助能力](/guide/helpers-packages) | `pkg/*`、`internal/pkg/*` 速查 |
| [生产运行手册](/ops/production-runbook) | systemd、Nginx、迁移、回滚 |
| [上线检查清单](/checklist) | 发布前打勾项（含 Worker、网关） |

## 其它能力（索引）

- **事件 / 通知**：`pkg/eventbus`、`pkg/notify` 与 [platform](/guide/platform) 的 Outbox、通知驱动章节。
- **雪花 ID**：`internal/pkg/snowflake` + 配置 `snowflake.node`，见 [配置详解](/guide/configuration-advanced)。
- **RBAC 实操**：见专页 [RBAC 与权限](/guide/rbac-and-permissions)。

## 与「官方级」文档的差距说明

- 本手册与 **OpenAPI/Swagger 生成物**（`docs/swagger.*`）互补：HTTP 字段级契约以 Swagger 为准。
- **业务域**（订单、支付等）需你们自建章节或 Wiki；脚手架只提供通用模式。
- 各指南页会逐步补充 **可复制命令 / 代码块**；新增文档时请遵守 **[文档贡献规范](/meta/docs-maintenance)**（含「尽量附可复制示例」）。
