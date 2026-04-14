# gin-scaffold

企业级 Go 脚手架（Gin + GORM + Redis + Asynq + Swagger + 可观测性）。

## 快速开始

```bash
go mod tidy
go build -o bin/server ./cmd/server
./bin/server server --env dev
```

## Docker 一键启动

```bash
docker-compose up -d
```

## 核心访问地址

- `http://localhost:8080/health`
- `http://localhost:8080/swagger/index.html`
- `http://localhost:8080/metrics`

## 单元测试

### 目录说明

- 常规包内测试：`**/*_test.go`
- 示例化单测目录：`tests/unit`

### 安装测试依赖

```bash
go get github.com/stretchr/testify@latest
go mod tidy
```

### 运行单测

```bash
go test ./tests/unit/...
```

或运行全量测试：

```bash
go test ./...
```

## Makefile 命令

可使用以下统一命令：

```bash
make tidy
make build
make run ENV=dev
make run-worker ENV=dev
make migrate-up DRIVER=mysql DSN="<database_dsn>"
make migrate-down DRIVER=mysql DSN="<database_dsn>"
make test-unit
make test
make swagger
```

> Windows 若未安装 `make`，可继续使用 README 中对应的 `go` 命令直接执行。

## Windows 命令入口

项目提供了 PowerShell 版统一命令：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\make.ps1 -Target tidy
powershell -ExecutionPolicy Bypass -File .\scripts\make.ps1 -Target build
powershell -ExecutionPolicy Bypass -File .\scripts\make.ps1 -Target run -Env dev
powershell -ExecutionPolicy Bypass -File .\scripts\make.ps1 -Target migrate-up -Driver mysql -Dsn "<database_dsn>"
powershell -ExecutionPolicy Bypass -File .\scripts\make.ps1 -Target test-unit
powershell -ExecutionPolicy Bypass -File .\scripts\make.ps1 -Target test
```

更多示例见 `Makefile.win`。

## 生产化增强已内置

- CI 工作流：`.github/workflows/ci.yml`
- 数据库迁移命令（Gormigrate）：`cmd/migrate`
- JWT 刷新：`/api/v1/auth/refresh`
- JWT 刷新轮换与重放防护：refresh token 单次使用（基于 Redis jti）
- JWT 吊销（黑名单）：`/api/v1/auth/logout`
- 管理端 RBAC：`/api/v1/admin/*` 要求 `role=admin` + `db:ping` 权限
- 环境变量模板：`.env.example`

## 配置文件是否提交到 Git

建议策略如下：

- **应提交**：
  - `configs/app.yaml`（dev 默认模板）
  - `configs/app.test.yaml`（test 模板）
  - `configs/app.prod.yaml`（prod 模板，仅占位符，不放真实密钥）
  - `configs/app.local.yaml.example`（本地覆盖模板）
- **不应提交**：
  - `configs/app.local.yaml`（本地私有配置）
  - `.env`（真实密钥/连接串）

仓库已默认忽略：`configs/app.local.yaml`。
