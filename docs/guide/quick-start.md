# 快速开始

## 1) 环境准备

- Go `1.22+`
- MySQL 或 PostgreSQL
- Redis

## 2) 拉取代码与依赖

```bash
git clone <your_repo_url>
cd gin-scaffold
go mod tidy
```

## 3) 准备 `.env.local`

先生成可逆加密密钥（可直接粘贴到 `.env.local`）：

```bash
go run ./cmd/artisan key:generate
```

```text
APP_ENV=dev
DB_DSN=root:root@tcp(127.0.0.1:3306)/gin_scaffold?charset=utf8mb4&parseTime=True
TIME_ZONE=UTC
HTTP_READ_TIMEOUT_SEC=30
HTTP_READ_HEADER_TIMEOUT_SEC=10
HTTP_WRITE_TIMEOUT_SEC=30
HTTP_IDLE_TIMEOUT_SEC=120
HTTP_SHUTDOWN_TIMEOUT_SEC=10
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
JWT_SECRET=replace-with-your-own-secret
ENCRYPTION_KEY=base64:replace-with-generated-key
TENANT_ENABLED=false
TENANT_HEADER=X-Tenant-ID
TENANT_DEFAULT_ID=default
```

## 4) 执行迁移

```bash
go run ./cmd/migrate up --env dev
go run ./cmd/migrate seed up --env dev
```

## 5) 启动服务

```bash
go run ./cmd/server server --env dev
go run ./cmd/server worker --env dev
```

若你需要在 `test/prod` 环境临时使用 `.env.test` / `.env.prod` 做联调，可先设置：

```bash
export LOAD_DOTENV_NON_DEV=true
```

Windows PowerShell:

```powershell
$env:LOAD_DOTENV_NON_DEV = "true"
```

## 6) 验证接口

- `http://localhost:8080/livez`
- `http://localhost:8080/readyz`
- `http://localhost:8080/swagger/index.html`（需在 `configs` 中 `http.swagger_enabled: true`，开发模板默认开启）

命令行示例（端口以你本地为准；种子用户口令见迁移 `seed_users` 内注释）：

```bash
curl -sS "http://127.0.0.1:8080/livez"

curl -sS -X POST "http://127.0.0.1:8080/api/v1/client/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123456"}'
```

## 7) 常见问题与排查

- `readyz` 失败：优先检查数据库/Redis 连接与 `DB_DSN`、`REDIS_ADDR`。
- 登录 401：确认 seed 用户口令与环境一致，并检查 `JWT_SECRET` 是否变化。
- worker 没消费任务：确认已启动 `go run ./cmd/server worker --env dev`。
- 非 dev 环境没读取 `.env.*`：确认设置了 `LOAD_DOTENV_NON_DEV=true`。

## 8) 变更记录

- 版本更新说明见：`docs/changelog.md`

## 9) 多租户启用（可选）

```bash
export TENANT_ENABLED=true
export TENANT_HEADER=X-Tenant-ID
export TENANT_DEFAULT_ID=default
```

Windows PowerShell:

```powershell
$env:TENANT_ENABLED = "true"
$env:TENANT_HEADER = "X-Tenant-ID"
$env:TENANT_DEFAULT_ID = "default"
```
