# gin-scaffold 文档中心

企业级 Go 脚手架（Gin + GORM + Redis + Asynq + Swagger + 可观测性）。

适用定位：**中小团队可生产使用**，强调稳定、可维护、可观测与快速迭代。

## 推荐入口

- **[开发手册（总览索引）](/guide/handbook)**：按主题串联路由、中间件、配置、队列、迁移、测试、安全等，**尽量只靠文档即可上手开发**。

## 你可以从这里开始

- 新同学：[新人入门](/guide/onboarding) → [快速开始](/guide/quick-start) → [目录结构](/guide/directory-structure) → [路由](/guide/routing)
- 运维同学：[生产运行手册](/ops/production-runbook) → [上线检查清单](/checklist)
- 测试同学：[测试指南](/guide/testing-guide) 与 [测试同学路径](/paths/testing)
- 按角色走：[开发同学](/paths/developer) / [运维同学](/paths/operations) / [测试同学](/paths/testing)

## 文档导航（精选）

- [项目简介](/guide/introduction)
- [配置说明](/guide/configuration) · [配置详解（全量键）](/guide/configuration-advanced) · [环境变量](/guide/environment-variables)
- [平台横切能力](/guide/platform)
- [命令系统](/guide/commands) · [代码生成 CRUD](/guide/codegen)
- [数据库迁移与填充](/guide/database-and-migrations) · [GORM 实践](/guide/database-patterns)
- [异步队列 Asynq](/guide/queues-asynq) · [定时任务](/guide/scheduler)
- [文件存储](/guide/file-storage) · [SSE / WebSocket](/guide/realtime-sse-websocket)
- [安全实践](/guide/security-practices) · [RBAC 与权限](/guide/rbac-and-permissions) · [管理端 API 总览](/guide/admin-api-overview) · [错误与响应](/guide/error-handling)
- [日志与可观测](/guide/logging-observability) · [出站 HTTP](/guide/outbound-httpclient)
- [文档贡献规范](/meta/docs-maintenance)
- 代码贡献、漏洞披露与许可证：仓库根目录 `CONTRIBUTING.md`、`SECURITY.md`、`LICENSE`（与 CI、Dependabot、PR 模板、`scripts/quality.sh` 配套）
