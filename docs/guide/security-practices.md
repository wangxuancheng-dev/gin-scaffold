# 安全实践

## 认证与授权

- **JWT**：密钥长度与轮换策略由运维负责；刷新令牌逻辑见客户端 `auth` 路由与 `internal/pkg/jwt`。
- **RBAC**：管理端接口以 `RequirePermission` 为准；权限与菜单由 **seed 迁移** 初始化。实操见 **[RBAC 与权限](/guide/rbac-and-permissions)**。
- **多租户**：生产务必 `tenant.enabled=true`，由网关统一注入 `X-Tenant-ID`，见 [platform](/guide/platform)。

## 暴露面控制

- **Swagger**：`http.swagger_enabled`，生产默认关闭；见 [配置说明](/guide/configuration)。
- **Prometheus**：`metrics.allowed_networks` 按 TCP 源 IP 限制；公网抓取请置空并只靠网关。
- **调试路由**：`/debug/panic` 仅 `debug` 且回环 IP。

## 数据与密钥

- 使用 **`scripts/deploy/check-prod-env.sh`** 做上线前自检。
- `.env.prod` 权限 `600`；不入库真实密钥到 Git。

## 请求与文件

- **请求体上限**：`http.max_body_bytes`。
- **上传**：扩展名 + MIME 白名单 + 大小上限（`storage.*`），见 [文件存储](/guide/file-storage)。
- **下载响应头**：对象 key 会写入 `Content-Disposition` 文件名，已做引号/换行剥离（`pkg/strutil.AttachmentFilename`）；仍应避免把不可信字符串直接拼进其它自定义头。

## 定时任务与 WebSocket

- **定时任务命令**：`artisan …` 走应用内注册表；其它命令历史上通过 `sh -c` / `cmd /C` 执行，等价于**可远程触发的 RCE**（能改 DB 任务的人即可执行）。生产配置 **`scheduler.shell_commands_enabled: false`**（默认），仅允许 `artisan`；确需 shell 时再显式打开并收紧谁能改任务。见 [定时任务中心](/guide/scheduler)。
- **WebSocket**：`/api/v1/client/ws` 在 **JWT 保护组** 内，身份以 token 为准；`CheckOrigin` 与 **`cors.allow_origins`** 对齐；`allow_origins` 含 `*` 或未配置时仍偏宽松，生产请列出明确前端源。见 [SSE/WebSocket](/guide/realtime-sse-websocket)。

## 业务安全横切

- **审计**：可选记录写操作元数据（不含 body），见 [platform](/guide/platform)。
- **幂等**：可选对 JSON POST + `X-Idempotency-Key` 生效。
- **登录防爆破**：`platform.login_security`（Redis）。

## 合规清单

完整打勾项见 **[上线检查清单](/checklist)**。
