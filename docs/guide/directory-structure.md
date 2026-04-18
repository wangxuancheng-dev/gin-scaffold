# 目录结构与分层

## 顶层一览

| 路径 | 职责 |
|------|------|
| `cmd/server` | HTTP 入口与 **Asynq Worker** 子命令（同一二进制） |
| `cmd/migrate` | 数据库结构迁移与 **seed** 迁移 |
| `cmd/gen` | CRUD / 代码生成 |
| `cmd/artisan` | 应用内控制台命令（`list` / `run` / `make:command` / `queue:failed` 等） |
| `configs/` | 多环境 YAML；敏感项用环境变量占位 |
| `config/` | Viper 加载、环境变量绑定、`Validate()` fail-fast |
| `routes/` | Gin 引擎构建、全局中间件、`/api/v1` 注册 |
| `routes/adminroutes/` | 管理端路由按模块拆分 |
| `middleware/` | 可复用 HTTP 中间件 |
| `api/handler/` | Controller：入参绑定、调用 service、统一错误出口 |
| `api/request/` | 请求 DTO（`binding` 标签） |
| `api/response/` | 统一 JSON 信封 |
| `internal/service/` | 业务编排；`port/` 下为接口便于测试 |
| `internal/dao/` | 数据访问；与 GORM、租户 scope 配合 |
| `internal/model/` | 领域模型 / 表映射 |
| `internal/app/bootstrap/` | 进程启动：DB、Redis、队列、调度器、路由组装 |
| `internal/job/` | Asynq 任务类型、handler、`scheduler`（DB Cron） |
| `pkg/` | 可对外复用的库级代码（logger、redis、storage、cache…） |
| `migrations/` | 按驱动分目录的 SQL（`mysql` / `postgres`） |
| `tests/unit`、`tests/integration` | 单元与集成测试 |
| `docs/` | VitePress 文档站 |
| `deploy/` | systemd、Nginx 示例、环境自检脚本 |

## 依赖方向（约定）

```text
api/handler → internal/service/port → internal/service → internal/dao → internal/model
                ↘ pkg/* 、 config
```

- **Handler 不写 SQL**；复杂查询放在 DAO。
- **跨模块复用**优先沉到 `pkg/` 或 `internal/pkg/`（如 `errcode`、`jwt`、`tenant`）。

## 新增业务模块的推荐顺序

1. `migrations/` 建表 + seed 权限（若需后台菜单）。
2. `internal/model` + `dao` + `service` + `port`。
3. `api/request` + `handler` + `response`（若有 VO）。
4. `routes/adminroutes`（或 `client_router`）注册路由。
5. `internal/app/bootstrap` 里 **构造依赖并传入** `routes.Build` 的 `Options`（与现有 User/Menu 等并列）。
6. 补 Swagger 注释后执行 `swag init`（见 [命令系统](/guide/commands) 与 CI 校验说明）。
