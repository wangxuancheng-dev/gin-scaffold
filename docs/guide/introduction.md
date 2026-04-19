# 项目简介

系统化的开发说明见 **[开发手册](/guide/handbook)**。

`gin-scaffold` 是一套面向业务落地的 Go 服务端脚手架，核心目标是：

- 快速启动：开箱即用的 API、迁移、任务、日志、鉴权能力
- 生产可用：基础可观测、超时、限流、恢复、检查点齐全
- 结构清晰：分层明确，易于团队协作和后续扩展

## 技术栈

- Web：Gin
- ORM：GORM
- DB：MySQL / PostgreSQL
- Cache：Redis
- Queue：Asynq
- Docs：Swagger + VitePress
- Logging：Zap + 按需轮转
- Trace/Metrics：OpenTelemetry + Prometheus

## 目录概览

```text
cmd/                # 启动命令（server/migrate/gen/artisan）
config/             # 配置加载与校验
configs/            # 各环境配置模板
internal/           # 核心业务（dao/service/job/console 等）
api/                # handler/request/response
routes/             # 路由注册
migrations/         # DB 迁移（schema/seed）
docs/               # 文档中心（VitePress）
```

## 设计原则

- 默认可运行，关键能力可配置
- 配置错误快速失败（fail fast）
- 功能优先服务中小团队生产场景，避免过度平台化

## 最小可运行验证（示例）

更完整的步骤见 **[快速开始](/guide/quick-start)**。假设已完成迁移且本机 `8080` 起服务：

```bash
# 健康检查（无需鉴权）
curl -sS "http://127.0.0.1:8080/livez"

# 客户端登录（种子用户见迁移 seed；默认口令见 seed 注释，首登后请修改）
curl -sS -X POST "http://127.0.0.1:8080/api/v1/client/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123456"}'
```

响应 `data` 中含 `access_token`，管理端接口在 Header 中携带：`Authorization: Bearer <token>`（需对应 RBAC 权限）。
