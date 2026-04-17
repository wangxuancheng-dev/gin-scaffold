# 代码生成（CRUD）

使用 `cmd/gen` 快速生成一个后台 CRUD 模块骨架。

## 命令

```bash
go run ./cmd/gen crud --module order
go run ./cmd/gen crud --module order --field title:string:required,max=64 --field amount:int:min=0 --field note:string?:max=255 --field status:string:oneof=draft published,default=draft
```

可选参数：

- `--template`：`full`（默认）或 `simple`
- `--table`：指定表名（默认 `<module>s`）
- `--field`：字段定义（可重复），格式 `name:type[:validate]`，支持 `string|int|int64|bool|float64`
  - 类型后缀 `?` 表示 create 可选字段（例如 `note:string?`）
  - `validate` 会原样写入 `binding`，update 自动加 `omitempty,`
  - 支持 `default=<value>`，会写入 migration SQL 的列默认值（如 `status:string:oneof=draft published,default=draft`）
- `--force`：覆盖已存在文件
- `--no-wire`：只生成文件，不自动注入路由/bootstrap
- `--dry-run`：仅预览将生成的文件
- `--preview-file`：输出预览 markdown（适合 code review）
- `--preview-full`：配合 `--preview-file` 输出完整文件内容（默认会截断长内容）
- `--out-dir`：指定生成输出目录（非项目根目录时需配合 `--no-wire`）

## 生成内容

- `internal/model/<module>.go`
- `internal/dao/<module>_dao.go`
- `internal/service/port/<module>_service.go`
- `internal/service/<module>_service.go`
- `api/request/admin/<module>_request.go`
- `api/handler/admin/<module>_handler.go`
- `routes/adminroutes/<module>_router.go`
- `migrations/mysql/schema/*_create_<table>.up.sql`
- `migrations/mysql/schema/*_create_<table>.down.sql`
- `migrations/mysql/seed/*_seed_<module>_permission.up.sql`
- `migrations/mysql/seed/*_seed_<module>_permission.down.sql`

`simple` 模板只生成代码骨架（不生成 migration/seed，也不自动 wiring）。
使用 `simple` 时会自动启用 `--no-wire`。

## 建议流程

1. 执行 `--dry-run` 先确认生成路径
2. 需要评审时可加 `--preview-file ./tmp/<module>-preview.md`
3. 运行生成命令并补全业务细节
4. 执行迁移：`go run ./cmd/migrate up --env dev`
5. 视需要补充菜单 seed 与权限分配

更多实战示例见：`docs/guide/codegen-walkthrough.md`
