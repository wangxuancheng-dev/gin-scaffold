# gin-scaffold

企业级 Go 脚手架（Gin + GORM + Redis + Asynq + Swagger + 可观测性），定位中小团队生产可用。

## 文档中心（VitePress）

项目文档已迁移到 VitePress，请优先阅读文档站：

- 本地文档首页：`http://localhost:5173`
- 文档入口文件：`docs/index.md`

### 本地运行文档

```bash
npm install
npm run docs:dev
```

### 构建静态文档

```bash
npm run docs:build
```

## 最小启动命令（TL;DR）

```bash
# 1) 迁移
go run ./cmd/migrate up --env dev

# 2) 启动服务
go run ./cmd/server server --env dev
```

## 常用入口链接（文档站）

- 快速开始：`/guide/quick-start`
- 配置说明：`/guide/configuration`
- 命令系统：`/guide/commands`
- 版本变更记录：`/changelog`
- 定时任务中心：`/guide/scheduler`
- 日志与可观测：`/guide/logging-observability`
- 生产运行手册：`/ops/production-runbook`
- 上线检查清单：`/ops/checklist`
- 按角色阅读路径：`/paths/developer`、`/paths/operations`、`/paths/testing`

## 仓库内关键目录

```text
cmd/                # server/migrate/gen/artisan
config/             # 配置加载 + 校验（fail fast）
configs/            # 多环境配置模板
internal/           # 业务核心层
api/                # handler/request/response
routes/             # 路由注册
migrations/         # schema + seed
docs/               # VitePress 文档站
```
