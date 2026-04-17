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

```env
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
TENANT_ENABLED=false
TENANT_HEADER=X-Tenant-ID
TENANT_DEFAULT_ID=default
```

## 4) 执行迁移

```bash
go run ./cmd/migrate up --env dev
```

## 5) 启动服务

```bash
go run ./cmd/server server --env dev
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
- `http://localhost:8080/swagger/index.html`

## 7) 变更记录

- 版本更新说明见：`docs/changelog.md`

## 8) 多租户启用（可选）

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
