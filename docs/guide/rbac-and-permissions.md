# RBAC 与权限

## 模型概要

- **角色**：JWT Claims 中带 `role`（如 `admin`、普通用户角色名）。
- **权限点**：字符串，如 `task:read`、`user:update`，与菜单、接口一一对应（由 seed SQL 初始化）。
- **超管**：`rbac.super_admin_user_id` 指定的用户 ID **跳过**角色与权限检查（见 `middleware/jwt_auth.go` 中 `isSuperAdminUser`）。生产务必设为真实超管账号 ID，环境变量 **`RBAC_SUPER_ADMIN_USER_ID`** 可覆盖。

## 管理端路由两层控制

1. **`RequireRoles("admin")`**：挂在 `/api/v1/admin` 整组（`routes/adminroutes/register.go`），先保证「管理员角色」。
2. **`RequirePermission("xxx")`**：挂在具体路由上，做细粒度授权。

示例（任务队列）：

```go
admin.GET("/task-queues/summary", middleware.RequirePermission("task:read"), h.Summary)
```

## 权限检查器注入

- 在 **`internal/app/bootstrap.InitServer`** 中：`middleware.SetPermissionChecker(authz.NewDBPermissionChecker(...))`。
- 若未注入，`RequirePermission` 会返回 **500**（`permission checker not configured`）。

## 新增接口时的 checklist

1. 在 `routes/adminroutes/*_router.go` 为路由增加 **`RequirePermission("your:action")`**。
2. 在 **`migrations/*/seed/`** 增加权限记录与（可选）菜单项，使角色可分配到该权限。
3. 本地执行 **`migrate seed up`** 或等价命令，验证非超管账号在 UI/接口上行为符合预期。

## 客户端 JWT

- 客户端路由使用 **`JWTAuth`**，不按 `RequirePermission` 细分（业务接口自行在 handler 内判断资源归属，可结合 `pkg/policy`）。

### 调用管理端接口（示例）

先登录获取 `access_token`（见 [快速开始](/guide/quick-start)），再请求需权限的接口：

```bash
curl -sS "http://127.0.0.1:8080/api/v1/admin/task-queues/summary" \
  -H "Authorization: Bearer <access_token>"
```

无对应权限时返回 **403**；超管用户 ID 由 `rbac.super_admin_user_id`（或 `RBAC_SUPER_ADMIN_USER_ID`）配置。

### Service 返回业务错误（与 `FailByError` 配合）

```go
import "gin-scaffold/internal/pkg/errcode"

// 资源不存在
return nil, errcode.New(errcode.NotFound, errcode.KeyNotFound)
```

Handler 侧用 `handler.FailByError` 将 `*errcode.BizError` 映射为 HTTP 状态与统一 JSON 信封，见 [错误与响应](/guide/error-handling)。

## 相关文档

- [路由与分组](/guide/routing)
- [安全实践](/guide/security-practices)
- [平台能力](/guide/platform)（租户与 RBAC 数据隔离说明）
