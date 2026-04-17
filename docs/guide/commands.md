# 命令系统

项目使用 Cobra 统一命令风格：

```bash
go run ./cmd/<tool> <subcommand> --flags
```

## Server

```bash
go run ./cmd/server server --env dev
go run ./cmd/server worker --env dev
```

## Migrate

```bash
go run ./cmd/migrate up --env dev
go run ./cmd/migrate down --env dev
go run ./cmd/migrate up --env dev --driver postgres --dsn "<dsn>"
```

## Gen

```bash
go run ./cmd/gen crud --module order --table orders
go run ./cmd/gen crud --module order --field title:string:required,max=64 --field amount:int:min=0 --field note:string?:max=255 --field status:string:oneof=draft published,default=draft
go run ./cmd/gen crud --module order --template simple --field title:string --no-wire
go run ./cmd/gen crud --module order --dry-run
go run ./cmd/gen crud --module order --dry-run --preview-file ./tmp/order-preview.md
go run ./cmd/gen crud --module order --dry-run --preview-file ./tmp/order-preview-full.md --preview-full
go run ./cmd/gen crud --module order --no-wire
go run ./cmd/gen crud --module order --no-wire --out-dir ./tmp/scaffold-preview
```

## Artisan

```bash
go run ./cmd/artisan list
go run ./cmd/artisan run ping
go run ./cmd/artisan make:command report:daily
# 查看死信（归档）任务
go run ./cmd/artisan queue:failed list --env dev
# 指定某个队列查看
go run ./cmd/artisan queue:failed list --env dev --queue critical
# 重试一条死信任务
go run ./cmd/artisan queue:failed retry <task_id> --env dev
# 指定队列重试
go run ./cmd/artisan queue:failed retry <task_id> --env dev --queue critical
```

### 自定义命令与计划任务联动

任务中心 `command` 支持 `artisan` 前缀，例如：

```bash
artisan ping
artisan report:daily --date=2026-04-16
```

这样任务调度与本地命令运行走同一套命令注册体系。

## Integration Test（关键链路）

先准备环境变量（服务需已启动）：

```bash
$env:INTEGRATION_BASE_URL="http://127.0.0.1:8080"
$env:INTEGRATION_ADMIN_USERNAME="admin"
$env:INTEGRATION_ADMIN_PASSWORD="admin123456"
```

执行集成测试：

```bash
go test -tags=integration ./tests/integration -v
```

一键本地依赖 + 集成测试（Windows / PowerShell）：

```powershell
.\scripts\integration.ps1 -Action all
# 或
.\scripts\make.ps1 -Target integration-all
```
