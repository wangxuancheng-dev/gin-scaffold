# 管理端 API 总览（后台）

所有管理端接口前缀：**`/api/v1/admin`**。均需：

- Header：`Authorization: Bearer <access_token>`
- 角色：JWT 中 `role` 为 **`admin`**
- 各路由另需 **`RequirePermission`** 中声明的权限码

**请求/响应字段级说明**以 **Swagger** 为准：本地或内网打开 **`/swagger/index.html`**（需 `http.swagger_enabled=true`）或阅读仓库内 `docs/swagger.json`。

## 模块与路由文件

| 模块 | 路由注册文件 | 典型权限前缀 |
|------|----------------|----------------|
| 用户 | `routes/adminroutes/user_router.go` | `user:*` 等（以 seed 为准） |
| 菜单 | `routes/adminroutes/menu_router.go` | 菜单/目录相关 |
| 运维 / 审计 / 导出 | `routes/adminroutes/ops_router.go` | `audit:*` 等 |
| 定时任务 | `routes/adminroutes/task_router.go` | `task:*` |
| 任务队列（Asynq） | `routes/adminroutes/task_queue_router.go` | `task:read` 等 |
| 系统参数 | `routes/adminroutes/system_setting_router.go` | `sys:config:*` |
| 公告（示例 CRUD） | `routes/adminroutes/announcement_router.go` | 以 seed 为准 |

入口聚合：`routes/adminroutes/register.go`。

## 调用示例（curl 思路）

1. 先 `POST /api/v1/client/auth/login` 取 `access_token`（见 Swagger）。
2. 管理请求带：`Authorization: Bearer ...`，多租户时加 **`X-Tenant-ID`**（与 [platform](/guide/platform) 一致）。

具体路径与 query/body 请用 Swagger 复制，避免文档与代码漂移。

## 与代码生成器的关系

- 使用 **`cmd/gen crud`** 生成的新模块：除路由文件外，记得在 **`register.go`** 中注册 `RouterGroup`，并补 **seed 权限**。
