# 开发同学阅读路径

目标：最快速度理解项目、完成开发、避免踩坑。

## 首选：开发手册

请先打开 **[开发手册（总览）](/guide/handbook)**，按其中表格跳转到具体章节（路由、中间件、缓存、队列、迁移、测试、安全等）。

## Day 0（30–60 分钟）

1. [新人入门：架构与常见问题](/guide/onboarding)
2. [项目简介](/guide/introduction)
3. [快速开始](/guide/quick-start)
4. [目录结构与分层](/guide/directory-structure)
5. [路由与分组](/guide/routing)
6. [配置说明](/guide/configuration)

## Day 1（开始写业务前）

1. [中间件参考](/guide/middleware-reference)
2. [错误与响应](/guide/error-handling)
3. [数据库迁移与填充](/guide/database-and-migrations) · [GORM 实践](/guide/database-patterns)
4. [异步队列（Asynq）](/guide/queues-asynq) 与 [定时任务中心](/guide/scheduler)
5. [命令系统](/guide/commands)（含 Artisan 与定时任务 `command` 联动）

## 按需深入

- [代码生成 CRUD](/guide/codegen) · [走读](/guide/codegen-walkthrough)
- [文件存储](/guide/file-storage)
- [缓存](/guide/caching) · [全局限流](/guide/rate-limiting) · [本地化](/guide/i18n)
- [SSE / WebSocket](/guide/realtime-sse-websocket)
- [配置详解](/guide/configuration-advanced) · [环境变量](/guide/environment-variables) · [平台横切](/guide/platform)
- [日志与可观测](/guide/logging-observability) · [出站 HTTP](/guide/outbound-httpclient)
- [安全实践](/guide/security-practices) · [RBAC 与权限](/guide/rbac-and-permissions) · [管理端 API 总览](/guide/admin-api-overview)
- [环境变量](/guide/environment-variables) · [测试指南](/guide/testing-guide)
- [常用包速查](/guide/helpers-packages)

## 开发中建议

- 本地统一使用 `dev` + `.env.local`（见快速开始）。
- 变更配置项时同步更新 `configs/*.yaml` 与相关文档章节。
- 新增 Artisan 命令时，可与 [定时任务](/guide/scheduler) 的 `artisan ...` 命令字段复用。

## 提交前自检

- `go test ./...`
- `bash scripts/check-handler-error-helper.sh`（或依赖 CI）
- Swagger 注释变更后执行 `swag init` 并与 `docs/swagger.*` 对齐
