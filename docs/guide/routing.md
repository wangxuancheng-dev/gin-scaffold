# 路由与分组

## 总入口

- 引擎在 `routes.Build`（`internal/routes/router.go`）中创建：全局中间件 → 健康检查 →（可选）Swagger → `registerAPIV1`。
- API 前缀：**`/api/v1`**。

## 客户端（无需登录 / 需登录）

源码：`internal/routes/client_router.go`。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/client/ping` | 连通性 |
| POST | `/api/v1/client/users` | 注册 |
| POST | `/api/v1/client/auth/login` | 登录 |
| POST | `/api/v1/client/auth/refresh` | 刷新令牌 |
| GET | `/api/v1/client/files/download` | 签名下载（查询参数验签） |
| GET | `/api/v1/client/ws` | WebSocket 演示（**需 JWT**，身份为 token 内用户） |
| GET | `/api/v1/client/sse/stream` | SSE 演示 |

以下路由挂在 **`JWTAuth` 之后** 的同一前缀组（`/api/v1/client`）：

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/client/users/:id` | 当前用户视角查询 |
| POST | `/api/v1/client/auth/logout` | 登出 |
| POST | `/api/v1/client/files/upload` | 上传 |
| POST | `/api/v1/client/files/presign` | 预签名 PUT |
| POST | `/api/v1/client/files/complete` | 完成上传 |
| GET | `/api/v1/client/files/url` | 签名访问 URL |

## 管理端（admin）

源码：`internal/routes/adminroutes/register.go`。

- 前缀：`/api/v1/admin`。
- 中间件顺序：`JWTAuth` → **`RequireRoles("admin")`** → 各子路由文件（`user_router.go`、`menu_router.go` …）。
- 细粒度接口权限使用 **`RequirePermission("xxx")`**（见各 `registerAdmin*Routes`）。

新增后台接口时：

1. 在对应 `*_router.go` 增加一行，并选择合适的 `RequirePermission`（需在 seed 中配置权限码与菜单）。
2. Handler 放在 `internal/api/handler/admin/`；请求体放 `internal/api/request/admin/`。

### 代码形态示例（管理端）

在 `internal/routes/adminroutes/register.go` 中已为 `admin` 路由组挂上 `JWTAuth` 与 `RequireRoles("admin")`。子模块文件（如 `user_router.go`）内为小写 `registerAdmin*Routes` 函数，例如：

```go
func registerAdminUserRoutes(admin *gin.RouterGroup, h *adminhandler.UserHandler) {
    admin.GET("/users", middleware.RequirePermission("user:read"), h.List)
    admin.POST("/users", middleware.RequirePermission("user:create"), h.Create)
}
```

`register.go` 中集中调用这些注册函数，保持「一组一个文件」便于 Code Review。

## 与 Swagger 的关系

- 路由以代码为准；注释用于 `swag init` 生成 `docs/swagger.*`。
- CI 会校验生成物与仓库一致，合并前需本地跑通生成命令。

## 常见扩展

- **新版本 API**：可复制 `registerAPIV1` 模式增加 `registerAPIV2`，或在大组下再分 `Group("/billing")`。
- **公开与鉴权分离**：参考 `client` 先注册无鉴权组，再注册 `Use(JWTAuth)` 的子组。
