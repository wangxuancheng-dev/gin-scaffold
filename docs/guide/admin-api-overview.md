# 管理端 API 总览（后台）

所有管理端接口前缀：**`/api/v1/admin`**。均需：

- Header：`Authorization: Bearer <access_token>`
- 角色：JWT 中 `role` 为 **`admin`**
- 各路由另需 **`RequirePermission`** 中声明的权限码

**请求/响应字段级说明**以 **Swagger** 为准：本地或内网打开 **`/swagger/index.html`**（需 `http.swagger_enabled=true`）或阅读仓库内 `docs/swagger.json`。

## 模块与路由文件

| 模块 | 路由注册文件 | 典型权限前缀 |
|------|----------------|----------------|
| 用户 | `internal/routes/adminroutes/user_router.go` | `user:*` 等（以 seed 为准） |
| 菜单 | `internal/routes/adminroutes/menu_router.go` | 菜单/目录相关 |
| 运维 / 审计 / 导出 | `internal/routes/adminroutes/ops_router.go` | `audit:*` 等 |
| 定时任务 | `internal/routes/adminroutes/task_router.go` | `task:*` |
| 任务队列（Asynq） | `internal/routes/adminroutes/task_queue_router.go` | `task:read` 等 |
| 系统参数 | `internal/routes/adminroutes/system_setting_router.go` | `sys:config:*` |
| 公告（示例 CRUD） | `internal/routes/adminroutes/announcement_router.go` | 以 seed 为准 |

入口聚合：`internal/routes/adminroutes/register.go`。

## 调用示例（curl）

1. 登录（种子用户与口令见迁移 `seed_users` 注释）：

```bash
curl -sS -X POST "http://127.0.0.1:8080/api/v1/client/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123456"}'
```

从响应 `data.access_token` 取出令牌。

2. 调用管理端（示例：队列摘要，需账号具备 `task:read` 等权限）：

```bash
TOKEN="<粘贴 access_token>"

curl -sS "http://127.0.0.1:8080/api/v1/admin/task-queues/summary" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-Tenant-ID: default"
```

多租户关闭时可省略 `X-Tenant-ID`。其余路径、query、body 以 **Swagger**（`/swagger/index.html`）为准，避免文档与代码漂移。

## 与代码生成器的关系

- 使用 **`cmd/gen crud`** 生成的新模块：除路由文件外，记得在 **`register.go`** 中注册 `RouterGroup`，并补 **seed 权限**。
