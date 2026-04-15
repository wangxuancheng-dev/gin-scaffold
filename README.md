# gin-scaffold

企业级 Go 脚手架（Gin + GORM + Redis + Asynq + Swagger + 可观测性）。

## 快速开始（新同学建议按这个走）

### 1) 环境准备

- Go：`1.22+`
- MySQL 或 PostgreSQL（本地可用 Docker）
- Redis

### 2) 拉代码并安装依赖

```bash
git clone <your_repo_url>
cd gin-scaffold
go mod tidy
```

### 3) 准备本地配置（推荐）

新建 `.env.local`（不要提交到 Git），至少配置：

```env
APP_ENV=dev
DB_DSN=root:root@tcp(127.0.0.1:3306)/gin_scaffold?charset=utf8mb4&parseTime=True
TIME_ZONE=UTC
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
JWT_SECRET=replace-with-your-own-secret
LOG_ROTATION_MODE=size
LOG_APP_ROTATION_MODE=
LOG_ACCESS_ROTATION_MODE=
LOG_ERROR_ROTATION_MODE=
```

> `dev` 环境会自动加载 `.env*` 系列文件，且不会覆盖你已设置的系统环境变量。
> `debug=true` 时会启用 Gin Debug 模式并输出 Gin 风格请求日志；`prod` 建议保持 `debug=false`。

### 4) 执行数据库迁移

```bash
go run ./cmd/migrate --env dev up
```

说明：

- `cmd/migrate` 在 `--env dev` 下会自动加载 `.env/.env.local`，并读取 `DB_DSN`
- MySQL 的 `DB_DSN` **不必再写 `loc=`**：`parseTime` 仍建议保留；驱动 **`Loc`** 与 **`TIME_ZONE` / `--time-zone`** 一致，由代码注入（与 HTTP 服务行为相同）
- 日志轮转支持全局 + 单文件覆盖：
  - 全局：`log.rotation_mode` / `LOG_ROTATION_MODE`，可选 `size`（默认）`daily` `none`
  - 单文件覆盖：`log.app_rotation_mode`、`log.access_rotation_mode`、`log.error_rotation_mode`（对应环境变量 `LOG_APP_ROTATION_MODE`、`LOG_ACCESS_ROTATION_MODE`、`LOG_ERROR_ROTATION_MODE`）
  - 当单文件模式为空时，回退使用全局模式；`daily` 按当前配置时区 0 点切换，文件名如 `app-2026-04-15.log`
- 支持自定义日志通道：
  - 在 `log.channels` 下按名称配置 `level`、`rotation_mode`（`file` 可选；可覆写 `max_size_mb/max_backups/max_age_days/compress`）
  - 代码中可用 `logger.Channel("<channel_name>", "<file_name>")` 动态指定文件名（未配置时自动回退主日志器）
- 支持 cron 示例任务（`robfig/cron/v3`）：
  - 配置在 `scheduler`：`enabled`、`with_seconds`、`log_retention_days`、`lock_enabled`、`lock_ttl_seconds`、`lock_prefix`
  - 调度规则来自数据库任务表（每条任务有独立 `spec` 与 `command`）
  - 执行日志写入 `scheduled_task_logs`，可通过后台接口查询；支持按保留天数自动清理
  - 多实例部署时通过 Redis 分布式锁防止同一任务被多台服务器重复执行（并带本机防重入）
- 如需临时覆盖可显式传 `--dsn`
- 数据库**会话时区**：未传 `--time-zone` 时读环境变量 **`TIME_ZONE`**（如 `UTC`、`Asia/Shanghai`、`+08:00`），否则默认 **`UTC`**（MySQL：`SET time_zone`；PostgreSQL：`SET TIME ZONE`），与迁移里 `NOW()` 一致；应用侧见 `configs` 的 **`db.time_zone`**（同样可用 **`TIME_ZONE`** 覆盖）。HTTP/Worker 启动时会把进程 **`time.Local`** 设成与该配置一致，**Gin 里没有单独时区开关**，普通 **`time.Now()`**、日志时间等与 **GORM 自动时间戳** 同一套时区语义
- 迁移目录默认按驱动自动选择：
  - MySQL: `migrations/mysql`（兼容回退 `migrations`）
  - PostgreSQL: `migrations/postgres`
  - 支持递归扫描子目录；建议按职责分目录：`schema/`（DDL）与 `seed/`（初始化数据）

PostgreSQL 示例（显式传参）：

```bash
go run ./cmd/migrate --env dev --driver postgres --dsn "<your_pg_dsn>" up
```

回滚上一次 migration（只回滚最后一步）：

```bash
go run ./cmd/migrate --env dev down
```

### 5) 启动服务

```bash
go run ./cmd/server --env dev
```

### 6) 本地验证

- 健康检查：`http://localhost:8080/livez`
- 就绪检查：`http://localhost:8080/readyz`
- Swagger：`http://localhost:8080/swagger/index.html`
- 调试 panic 验证（仅 `debug=true`）：`http://localhost:8080/debug/panic`

### 7) 常用导出接口示例

- CSV（默认）：
  - `GET /api/v1/admin/users/export?fields=id,username,nickname`
- XLSX：
  - `GET /api/v1/admin/users/export?export_format=xlsx&fields=id,username,role`
- 当前页导出：
  - `GET /api/v1/admin/users/export?export_scope=page&page=1&page_size=20`
- 大数据调优：
  - `GET /api/v1/admin/users/export?export_limit=1000000&export_batch_size=2000`

### 8) Windows 一键跑通（PowerShell）

如果你在 Windows 下开发，可直接执行下面命令：

```powershell
# 1) 安装依赖
go mod tidy

# 2) 准备本地配置（按需修改 DB/Redis/JWT）
@"
APP_ENV=dev
DB_DSN=root:root@tcp(127.0.0.1:3306)/gin_scaffold?charset=utf8mb4&parseTime=True
TIME_ZONE=UTC
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
JWT_SECRET=replace-with-your-own-secret
"@ | Set-Content -Path .env.local -Encoding UTF8

# 3) 执行迁移（MySQL 示例）
go run ./cmd/migrate --env dev up

# 4) 启动服务
go run ./cmd/server --env dev
```

启动后访问：

- `http://localhost:8080/livez`
- `http://localhost:8080/readyz`
- `http://localhost:8080/swagger/index.html`

## Docker 一键启动

```bash
docker-compose up -d
```

### Docker 零配置本地开发（推荐给新同学）

如果你本机没有安装 MySQL/Redis，可直接用 Docker 跑依赖：

```bash
# 1) 启动基础依赖（MySQL/Redis）
docker-compose up -d mysql redis

# 2) 配置本地环境变量（示例）
cp .env.example .env.local
# 按你的 docker-compose 实际端口和账号改 DB_DSN/REDIS_ADDR

# 3) 执行数据库迁移
go run ./cmd/migrate --env dev up

# 4) 启动服务
go run ./cmd/server --env dev
```

Windows（PowerShell）可用：

```powershell
docker-compose up -d mysql redis
go run ./cmd/migrate --env dev up
go run ./cmd/server --env dev
```

## 核心访问地址

- `http://localhost:8080/livez`
- `http://localhost:8080/readyz`
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

时间建议：MySQL 连接串不必再写 `loc=`，驱动 Loc 与 `TIME_ZONE` / `db.time_zone` 一致；存库与展示策略仍建议在业务层约定（常见为库内 UTC、接口再转本地）。

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
make run ENV=dev
# make run ENV=dev PROFILE=order
make run-worker ENV=dev
# make run-worker ENV=dev PROFILE=order
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
powershell -ExecutionPolicy Bypass -File .\scripts\make.ps1 -Target migrate-up -Dsn "<database_dsn>"
powershell -ExecutionPolicy Bypass -File .\scripts\make.ps1 -Target test-unit
powershell -ExecutionPolicy Bypass -File .\scripts\make.ps1 -Target test
```

更多示例见 `Makefile.win`。

## CRUD 自动生成（v1）

当前提供了一个基础生成器，适合快速起模块骨架（单表 + admin CRUD）：

```bash
go run ./cmd/gen crud --module order --table orders
```

可选参数：

- `--table`：自定义表名，默认 `<module>s`
- `--force`：覆盖已存在文件

会生成以下目录骨架（按当前分层约定）：

- `internal/model`
- `internal/dao`
- `internal/service`
- `internal/service/port`
- `api/request/admin`
- `api/handler/admin`
- `routes/admin_<module>_router.go`

生成后还需要你手动完成：

1. 补齐生成 handler 里的 request -> model 字段映射
2. 编写对应 migration SQL（建表/索引/约束）
3. 在 RBAC 表里增加模块权限并分配角色

> v2 已支持自动注入：会自动更新 `bootstrap`、`routes/router.go`、`routes/api_router.go`、`routes/admin_router.go`。

## 生产化增强已内置

- CI 工作流：`.github/workflows/ci.yml`
- 数据库迁移命令（Gormigrate）：`cmd/migrate`
- 客户端接口前缀：`/api/v1/client/*`
- 后台接口前缀：`/api/v1/admin/*`
- JWT 刷新：`/api/v1/client/auth/refresh`
- JWT 刷新轮换与重放防护：refresh token 单次使用（基于 Redis jti）
- JWT 吊销（黑名单）：`/api/v1/client/auth/logout`
- 管理端 RBAC：`/api/v1/admin/*` 要求 `role=admin` + `db:ping` 权限
- RBAC 权限来源：数据库 `roles` / `user_roles` / `role_permissions`（无配置兜底）
- 超管保护：`rbac.super_admin_user_id`（可由环境变量 `RBAC_SUPER_ADMIN_USER_ID` 覆盖）；该用户默认拥有全部权限且不允许删除
- 后台菜单可见性：`menus` + `role_menus`
- RBAC 数据表：`roles`、`user_roles`、`role_permissions`（见 `migrations/mysql/schema/202501011210_create_rbac.up.sql`）
- 用户管理权限矩阵：
  - `GET /api/v1/admin/users`、`GET /api/v1/admin/users/{id}`：`user:read`
  - `POST /api/v1/admin/users`：`user:create`
  - `PUT /api/v1/admin/users/{id}`：`user:update`
  - `DELETE /api/v1/admin/users/{id}`：`user:delete`
  - `GET /api/v1/admin/users/export`：`user:export`
- 任务中心权限矩阵：
  - `GET /api/v1/admin/tasks`、`GET /api/v1/admin/tasks/{id}/logs`：`task:read`
  - `POST /api/v1/admin/tasks`：`task:create`
  - `PUT /api/v1/admin/tasks/{id}`：`task:update`
  - `DELETE /api/v1/admin/tasks/{id}`：`task:delete`
  - `POST /api/v1/admin/tasks/{id}/toggle`：`task:toggle`
  - `POST /api/v1/admin/tasks/{id}/run`：`task:run`
- 管理员角色初始化：`migrations/mysql/seed/202501011230_seed_admin_role.up.sql`（按用户名 `admin` 绑定）
- 管理员账号初始化：`migrations/mysql/seed/202501011240_seed_admin_user.up.sql`（默认密码 `Admin@123456`，上线后立刻修改）
- 环境变量模板：`.env.example`
- `metrics.path` 与 `i18n` 配置项已接入运行时行为

## 前后台目录约定

- 客户端 handler：`api/handler/client`
- 后台 handler：`api/handler/admin`
- 共享用户服务接口：`internal/service/port`
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

> 生产发布前请先执行数据库迁移（详见运行手册“线上数据库迁移”小节），再进行服务重启与流量切换。
> 建议与 `server` 一起构建并上传 `migrate` 二进制（`go build -o bin/migrate ./cmd/migrate`），线上直接执行迁移命令，避免依赖 Go 运行环境。

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
- **不应提交**：
  - `.env.local`（本地私有配置）
  - `.env`（真实密钥/连接串，除非是纯示例）

仓库已默认忽略：`.env`（建议本地使用 `.env.local`）。

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
