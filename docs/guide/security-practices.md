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

## 最小可复制检查

生产环境常用快速核查：

```bash
# 1) swagger 应关闭（预期 404 或网关拦截）
curl -i "http://127.0.0.1:8080/swagger/index.html"

# 2) metrics 受限（从非白名单来源访问应失败）
curl -i "http://127.0.0.1:8080/metrics"

# 3) 关键配置核查（本地/CI）
bash ./scripts/check-security-baseline.sh .
```

## 定时任务与 WebSocket

- **定时任务命令**：`artisan …` 走应用内注册表；其它命令历史上通过 `sh -c` / `cmd /C` 执行，等价于**可远程触发的 RCE**（能改 DB 任务的人即可执行）。生产配置 **`scheduler.shell_commands_enabled: false`**（默认），仅允许 `artisan`；确需 shell 时再显式打开并收紧谁能改任务。见 [定时任务中心](/guide/scheduler)。
- **WebSocket**：`/api/v1/client/ws` 在 **JWT 保护组** 内，身份以 token 为准；`CheckOrigin` 与 **`cors.allow_origins`** 对齐；`allow_origins` 含 `*` 或未配置时仍偏宽松，生产请列出明确前端源。见 [SSE/WebSocket](/guide/realtime-sse-websocket)。

## 业务安全横切

- **审计**：可选记录写操作元数据（不含 body），见 [platform](/guide/platform)。
- **幂等**：可选对 JSON POST + `X-Idempotency-Key` 生效。
- **登录防爆破**：`platform.login_security`（Redis）。

## 合规清单

完整打勾项见 **[上线检查清单](/checklist)**。

## 常见问题与排查

- 生产仍可访问 Swagger：检查 `http.swagger_enabled=false` 是否被环境变量覆盖。
- `/metrics` 被公网抓取：检查 `metrics.allowed_networks` 与网关 ACL 是否一致。
- 登录防爆破未生效：确认 `platform.login_security.enabled=true` 且 Redis 可用。
- 定时任务可执行 shell：确认 `scheduler.shell_commands_enabled=false` 并限制任务修改权限。

## 生产配置片段（示例）

以下仅为常见收紧项，实际以你们 `configs/app.prod.yaml` 与网关策略为准。

```yaml
http:
  swagger_enabled: false

scheduler:
  shell_commands_enabled: false
```

```yaml
metrics:
  enabled: true
  path: /metrics
  allowed_networks:
    - 10.0.0.0/8
    - 127.0.0.1/32
```

`metrics.allowed_networks` 非空时按 **TCP 对端 IP** 校验（见 [配置说明](/guide/configuration)）；公网抓取请改为网关控制并酌情置空列表。
