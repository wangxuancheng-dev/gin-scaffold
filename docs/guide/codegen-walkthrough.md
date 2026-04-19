# Codegen 实战：从 0 到 1 生成 Announcement 模块

这份文档演示如何使用 `cmd/gen crud` 在项目里快速落地一个后台模块（以 `announcement` 为例）。

## 1) 生成模块骨架

```bash
go run ./cmd/gen crud \
  --module announcement \
  --template full \
  --field title:string:required,max=128 \
  --field content:string:required \
  --field status:string:oneof=draft published,default=draft
```

执行后会生成：

- model / dao / service / request / handler / route
- migration（建表，**仅** `migrations/mysql/schema/`）
- seed（权限，**仅** `migrations/mysql/seed/`）
- 自动 wiring（routes + bootstrap）

若使用 PostgreSQL，请在 `migrations/postgres/schema/`、`migrations/postgres/seed/` 中按同名时间戳补充等价 SQL（本仓库已提供与 MySQL 对齐的基线，可参考方言差异）。

## 2) 执行迁移

```bash
go run ./cmd/migrate up --env dev
go run ./cmd/migrate seed up --env dev
```

如果是测试环境：

```bash
go run ./cmd/migrate up --env test
go run ./cmd/migrate seed up --env test
```

## 3) 启动服务并验证路由

```bash
go run ./cmd/server server --env dev
go run ./cmd/server worker --env dev
```

建议用 Swagger 检查 `admin-announcement` 分组接口是否可见。

## 4) 权限与菜单

生成器会产出 announcement 的权限 seed（`announcement:read` / `announcement:write`）。

你可以：

1. 在 seed SQL 中为 `admin` 绑定权限
2. 按需新增菜单并绑定到 `role_menus`

## 5) 推荐收尾动作

1. 为 `announcement` 增加最小单测（至少 1 条 handler happy-path）
2. 补充接口字段说明（Swagger 注释）
3. 执行：

```bash
go test ./...
go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -o docs -d ./cmd/server,./api
```

## 常见问题

- **想先评审不落盘？**
  - 加 `--dry-run`，配合 `--preview-file ./tmp/preview.md`
- **想输出到临时目录？**
  - 使用 `--out-dir ./tmp/scaffold-preview --no-wire`
- **只要最小代码，不要 migration/seed？**
  - 使用 `--template simple`
