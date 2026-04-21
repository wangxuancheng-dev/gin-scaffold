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

## 高标准约束（建议作为代码评审基线）

- **封装边界**：`internal/*` 为业务核心与基础设施，`api/*`/`routes/*`/`middleware/*` 仅承担 HTTP 适配职责，不反向依赖具体 DAO。
- **错误出口统一**：Handler 层优先使用 `api/handler/error_helper.go` 的 `FailInvalidParam` / `FailNotFound` / `FailInternal`，避免散落 `response.FailHTTP(...)` 常量组合。
- **自动接线可失败不可静默**：代码生成自动 wiring 若锚点缺失应直接报错，禁止“看似成功但未真正接线”。
- **命名稳定性**：路由资源名、权限前缀、文件名统一基于 `snake_case module`，避免 `OrderItem -> orderitem` 这类不可读路径。

## 新增业务模块的推荐顺序

1. `migrations/` 建表 + seed 权限（若需后台菜单）。
2. `internal/model` + `dao` + `service` + `port`。
3. `api/request` + `handler` + `response`（若有 VO）。
4. `routes/adminroutes`（或 `client_router`）注册路由。
5. `internal/app/bootstrap` 里 **构造依赖并传入** `routes.Build` 的 `Options`（与现有 User/Menu 等并列）。
6. 补 Swagger 注释后执行 `swag init`（见 [命令系统](/guide/commands) 与 CI 校验说明）。

## 代码穿行示例（对照源码）

**Handler**：只做绑定、调 service、用统一错误出口与 `response.OK`（示例形态，与 `api/handler/admin/*` 一致；省略 `import`）：

```go
func (h *WidgetHandler) Get(c *gin.Context) {
    var uri struct {
        ID int64 `uri:"id" binding:"required,min=1"`
    }
    if err := c.ShouldBindUri(&uri); err != nil {
        handler.FailInvalidParam(c, err)
        return
    }
    row, err := h.svc.Get(c.Request.Context(), uri.ID)
    if err != nil {
        handler.FailInternal(c, err)
        return
    }
    response.OK(c, row)
}
```

**路由**（管理端 + 权限串，见 `routes/adminroutes/*.go`）：

```go
admin.GET("/widgets/:id", middleware.RequirePermission("widget:read"), h.Get)
```

**Bootstrap 接线**：在 `internal/app/bootstrap/bootstrap.go` 的 `InitServer` 中 `dao`/`service`/`handler` 与 `routes.Build(routes.Options{ ... })` 里为同一模块增加字段并传入（可对照 `AdminUser`、`AdminMenu` 的写法）。

**异步任务消费**：新 Asynq 类型除 `internal/job` 定义常量与 payload 外，必须在 **`InitWorker`** 里 `mux.Handle(...)` 注册，否则任务会一直 `pending`（见 [异步队列](/guide/queues-asynq)）。
