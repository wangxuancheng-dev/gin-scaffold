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
```

## Artisan

```bash
go run ./cmd/artisan list
go run ./cmd/artisan run ping
go run ./cmd/artisan make:command report:daily
# 查看死信（归档）任务
go run ./cmd/artisan queue:failed list --env dev
# 重试一条死信任务
go run ./cmd/artisan queue:failed retry <task_id> --env dev
```

### 自定义命令与计划任务联动

任务中心 `command` 支持 `artisan` 前缀，例如：

```bash
artisan ping
artisan report:daily --date=2026-04-16
```

这样任务调度与本地命令运行走同一套命令注册体系。
