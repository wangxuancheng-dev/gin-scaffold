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

## 本地开发配置（推荐）

当前在 `dev` 环境支持自动加载 `.env`（按环境与画像分层）：

- `.env`
- `.env.{env}`
- `.env.{env}.{profile}`（可选）
- `.env.local`
- `.env.{env}.local`

加载顺序按上面从前到后合并，且**不会覆盖已有系统环境变量**。  
因此多人开发时，建议每个人只维护自己的 `.env.local`，不改 `configs/app.yaml`。

`test/prod` 建议由 CI、容器或部署平台直接注入环境变量，不依赖 `.env` 文件。

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
make run ENV=dev PROFILE=order
make run-worker ENV=dev PROFILE=order
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
powershell -ExecutionPolicy Bypass -File .\scripts\make.ps1 -Target run -Env dev -Profile order
powershell -ExecutionPolicy Bypass -File .\scripts\make.ps1 -Target migrate-up -Driver mysql -Dsn "<database_dsn>"
powershell -ExecutionPolicy Bypass -File .\scripts\make.ps1 -Target test-unit
powershell -ExecutionPolicy Bypass -File .\scripts\make.ps1 -Target test
```

更多示例见 `Makefile.win`。

## 生产化增强已内置

- CI 工作流：`.github/workflows/ci.yml`
- 数据库迁移命令（Gormigrate）：`cmd/migrate`
- 客户端接口前缀：`/api/v1/client/*`
- 后台接口前缀：`/api/v1/admin/*`
- JWT 刷新：`/api/v1/client/auth/refresh`
- JWT 刷新轮换与重放防护：refresh token 单次使用（基于 Redis jti）
- JWT 吊销（黑名单）：`/api/v1/client/auth/logout`
- 管理端 RBAC：`/api/v1/admin/*` 要求 `role=admin` + `db:ping` 权限
- 环境变量模板：`.env.example`

## 前后台目录约定

- 客户端 handler：`api/handler/client`
- 后台 handler：`api/handler/admin`
- 共享用户服务接口：`api/handler/userapi`
- 客户端 request/response：`api/request/client`、`api/response/client`
- 后台 request：`api/request/admin`

## 生产部署与运维文档

- 生产运行手册（单机/小规模）：`docs/production-runbook.md`
- 上线前检查清单：`docs/checklist.md`
- systemd 服务模板：`deploy/systemd/gin-scaffold.service.example`
- Nginx 反向代理模板：`deploy/nginx/gin-scaffold.conf.example`
- Nginx 快速操作说明：`deploy/nginx/README.md`
- 生产环境变量模板：`deploy/.env.prod.example`
- 生产环境变量自检脚本：`scripts/deploy/check-prod-env.sh`

推荐生产部署路径：

1. 按运行手册准备 `configs/app.prod.yaml` 和服务器本地 `.env.prod`
2. 安装 systemd 服务模板并启动
3. 用 `/health` 与核心接口完成发布后巡检

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

## 多套生产系统（Profile）

当你有多套生产系统时，建议使用 `profile` 覆盖层：

- 基础层：`app.yaml`
- 环境层：`app.prod.yaml`
- 画像层：`app.prod.{profile}.yaml`

加载顺序为：基础 -> 环境 -> 画像 -> 环境变量（后者覆盖前者）。

示例：

```bash
./bin/server server --env prod --profile order
```

可参考示例文件：`configs/app.prod.order.yaml.example`。
