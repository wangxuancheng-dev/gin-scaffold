# 平台横切能力（审计 / 幂等 / 缓存 / 事件 / 通知）

与具体业务模块解耦的「框架向」能力，默认关闭或低侵入，可按需在 YAML 中打开。

## 配置入口

顶层 `platform:`，环境变量见 `config/loader.go` 中 `platform.*` 绑定（如 `PLATFORM_AUDIT_ENABLED`）。

## 审计日志（`platform.audit`）

- `enabled: true` 时，对 `POST`/`PUT`/`PATCH`/`DELETE` 异步写入表 `audit_logs`（需执行迁移 `202504171400_create_audit_logs`）。
- 记录：`request_id`、用户（JWT 或 anonymous）、路径、查询串、状态码、耗时、客户端 IP；**不落库请求体**。
- `/livez`、`/readyz`、`/health`、`/swagger`、`/metrics`、`/debug` 等路径自动跳过。
- 导出接口（`/api/v1/admin/audit-logs/export`）默认导出最近 `export_default_days` 天，且时间窗不得超过 `export_max_days` 天。
- 查询/导出权限：`audit:read`、`audit:export`（与审计菜单一并写入种子：`202501011226_seed_role_permissions`、`202501011227_seed_menus`、`202501011228_seed_role_menus`；执行 `go run ./cmd/migrate seed up`）。
- 导出响应头：`X-Export-Count`（本次导出行数）、`X-Export-Window`（实际导出时间窗，RFC3339/RFC3339）。
- 大数据量建议使用异步导出任务：
  - 创建：`POST /api/v1/admin/audit-logs/export/tasks`
  - 查询状态：`GET /api/v1/admin/audit-logs/export/tasks/{task_id}`
  - 下载结果：`GET /api/v1/admin/audit-logs/export/tasks/{task_id}/download`
  - 状态为 `success` 时，状态接口会返回 `download_url`（可直接下载）与 `filter`（任务筛选摘要）。
  - 任务固定投递到 Asynq `low` 队列，避免影响在线请求。

## 幂等（`platform.idempotency`）

- `enabled: true` 且请求带 `X-Idempotency-Key` 时，对 **`Content-Length` 已知的 JSON POST**、路径前缀 `/api/v1/` 的请求做 Redis 缓存。
- 成功（2xx）且 `Content-Type` 为 JSON 的响应体会被缓存，重复请求直接重放；并发相同指纹返回 **409**。
- `multipart`、过大请求体、未知 `Content-Length` 会跳过幂等逻辑。
- 指纹包含：用户维度（JWT `uid` 或 `anon`）、幂等键、路径、**原始请求体**。

## 缓存前缀（`platform.cache`）

- `pkg/cache` 的 `NewFromConfig()` 使用 `key_prefix` 拼接业务 Redis 键（默认 `app:`）。

## 事件总线（`pkg/eventbus`）

- 进程内**同步**派发；耗时逻辑请投递 Asynq。
- 在 `internal/app/platform.Init` 中重置默认总线；业务可 `eventbus.Default().On("name", handler)` 订阅。
- 示例：`user.registered` 在用户注册事务内写入 Outbox（见 `UserService.Register`），由 Outbox 分发器发布到本进程总线。

## 事务 Outbox（`outbox`）

- 目标：保证“数据库写入成功后，事件最终一定可分发”，避免直接异步调用导致的丢消息。
- 表：`outbox_events`（迁移：`202604172130_create_outbox_events`）。
- 运行机制：
  - 业务在事务内写入 outbox 记录（`status=pending`）；客户端注册 `POST /api/v1/client/users` 成功时同事务写入 `user.registered`（`outbox.enabled=true` 且已迁移 `outbox_events` 时生效）。
  - 后台分发器按 `poll_interval_sec` 扫描待处理记录；`publish_mode` 为 `eventbus`（默认）时发布到进程内 `eventbus`，为 `http` 时向 `http_url` POST JSON（可配 `http_hmac_secret` 生成 `X-Outbox-Signature`）。
  - 成功后标记 `published`；失败按 `retry_backoff_sec` 退避重试，超过 `max_attempts` 标记 `dead`。
- 配置项（`outbox`）：
  - `enabled`
  - `poll_interval_sec`
  - `batch_size`
  - `max_attempts`
  - `retry_backoff_sec`
  - `publish_mode`（`eventbus` | `http`）
  - `http_url` / `http_hmac_secret` / `http_timeout_ms`（`publish_mode=http` 时）

## 通知（`platform.notify` + `pkg/notify`）

- `driver: log`：写入应用日志（默认）。
- `driver: noop`：丢弃。
- `driver: smtp`：发邮件（需 `platform.notify.smtp.*`；`Meta["to"]` 可覆盖 `to_default`）。
- `driver: webhook`：向 `platform.notify.webhook.url` POST JSON（可选 `hmac_secret` → `X-Notify-Signature`）。
- 逗号分隔并行投递，例如 `smtp,webhook`。
- 示例：注册成功后 `notify.Default().Notify(...)`。

## 全局限流（`limiter`）

- `mode: memory`（默认）：进程内令牌桶。
- `mode: redis`：多实例共享的固定窗口计数（需 `window_sec`；可选 `redis_key_prefix`，否则回退 `platform.cache.key_prefix`）。

## 登录防爆破（`platform.login_security`）

- `enabled: true` 时依赖 Redis：窗口内失败次数达阈值后锁定 `lockout_sec`；成功登录会清理计数。
- 需在路由启用 `ClientIPContext`（已挂在全局中间件）以便 `UserService.Login` / `LoginWithRefresh` 读取 IP。

## 策略辅助（`pkg/policy`）

- `policy.SameUser(actorID, ownerID)`：判断资源是否属于当前用户，可与 RBAC 组合使用。

## 系统参数（System Settings）

- 后台接口（权限）：
  - `GET /api/v1/admin/system-settings`（`sys:config:read`）
  - `GET /api/v1/admin/system-settings/{id}`（`sys:config:read`）
  - `GET /api/v1/admin/system-settings/{id}/history`（`sys:config:read`）
  - `POST /api/v1/admin/system-settings`（`sys:config:write`）
  - `PUT /api/v1/admin/system-settings/{id}`（`sys:config:write`）
  - `DELETE /api/v1/admin/system-settings/{id}`（`sys:config:write`）
  - `POST /api/v1/admin/system-settings/{id}/publish`（`sys:config:publish`）
  - `POST /api/v1/admin/system-settings/{id}/rollback`（`sys:config:rollback`）
- 数据表：`system_settings`（包含草稿/发布字段与租户字段，迁移：`202504171500_create_system_settings`）。
- 变更历史表：`system_setting_histories`（迁移：`202604171030_create_system_setting_histories`）。
- 相关权限与菜单种子已合并进：`202501011226_seed_role_permissions`、`202501011227_seed_menus`、`202501011228_seed_role_menus`（执行 `go run ./cmd/migrate seed up`）。
- 业务读取建议使用 `pkg/settings`：
  - `settings.GetString(ctx, "your.key")`
  - `settings.GetInt64(ctx, "your.key")`
  - `settings.GetBool(ctx, "your.key")`
  - 内置短 TTL 缓存，减少高频读取数据库压力。
  - 默认只读取 `is_published=1` 的已发布配置。

## 多租户（Tenant Foundation）

- 配置入口：`tenant.enabled`、`tenant.header`、`tenant.default_id`。
- 解析优先级（启用后）：
  - 先读请求头（默认 `X-Tenant-ID`）
  - 为空则读 JWT `tenant_id`
  - 仍为空则回退 `tenant.default_id`（默认 `default`）
- 目前已租户隔离的数据主链路：
  - 用户与角色权限（`users` / `user_roles` / `role_permissions` / `roles`）
  - 菜单与角色菜单（`menus` / `role_menus`；`menus.parent_id` 树形，`GET /api/v1/admin/menus` 与 `/menus/catalog` 返回 `data.tree`）
  - 定时任务与任务日志（`scheduled_tasks` / `scheduled_task_logs`）
  - 审计日志（`audit_logs`）
  - 系统参数（`system_settings` / `system_setting_histories`）
- 关键约束：
  - 默认 seed 全部写入租户 `default`
  - 权限判断（RBAC）按当前租户查询，不跨租户复用
  - 建议生产环境启用 `tenant.enabled=true`，并统一由网关注入 `X-Tenant-ID`

## 用户异步导出（仅任务模式）

用户导出已统一为异步任务接口，不再提供同步 `GET /api/v1/admin/users/export`。

### 1) 创建任务

- 接口：`POST /api/v1/admin/users/export/tasks`
- 权限：`user:export`
- 可选筛选：`username`、`nickname`、`fields`（如 `id,username,nickname,created_at,role`）

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/admin/users/export/tasks?username=ali&fields=id,username,role" \
  -H "Authorization: Bearer <admin-jwt>"
```

返回示例（节选）：

```json
{
  "code": 200,
  "data": {
    "task_id": "e7d6e1e5-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "state": "queued",
    "filter": "file_type=csv&username~=ali&fields=id,username,role&with_role=true"
  }
}
```

### 2) 轮询任务状态

- 接口：`GET /api/v1/admin/users/export/tasks/{task_id}`
- 当 `state=success` 且 `is_ready=true` 时，响应会返回 `download_url`

```bash
curl "http://127.0.0.1:8080/api/v1/admin/users/export/tasks/e7d6e1e5-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  -H "Authorization: Bearer <admin-jwt>"
```

返回示例（节选）：

```json
{
  "code": 200,
  "data": {
    "task_id": "e7d6e1e5-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "state": "success",
    "is_ready": true,
    "progress_rows": 12345,
    "download_url": "http://127.0.0.1:8080/api/v1/admin/users/export/tasks/e7d6e1e5-xxxx-xxxx-xxxx-xxxxxxxxxxxx/download"
  }
}
```

### 3) 下载结果文件

- 接口：`GET /api/v1/admin/users/export/tasks/{task_id}/download`
- 当前导出文件类型固定为 CSV（后台 low 队列异步生成）

```bash
curl -L "http://127.0.0.1:8080/api/v1/admin/users/export/tasks/e7d6e1e5-xxxx-xxxx-xxxx-xxxxxxxxxxxx/download" \
  -H "Authorization: Bearer <admin-jwt>" \
  -o users_export.csv
```
