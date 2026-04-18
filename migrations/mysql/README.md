# MySQL Migrations Guide

本目录按职责拆分为两个子目录：

- `schema/`：结构变更（DDL），如 `CREATE TABLE`、`ALTER TABLE`、索引与约束。
- `seed/`：初始化与演示数据（DML），如默认角色、菜单、管理员账号。

## 执行规则

- 迁移入口为 `cmd/migrate`，默认根目录为 `migrations/mysql`。
- **结构**与 **种子数据** 分开执行（仍共用 gormigrate 的 `migrations` 表记录已执行的 ID）：
  - `go run ./cmd/migrate up --env dev`：只扫描并执行 `schema/**/*.up.sql`（存在 `schema/` 时）；若没有 `schema/` 子目录则兼容旧布局，扫描整个根目录。
  - `go run ./cmd/migrate seed up --env dev`：只扫描并执行 `seed/**/*.up.sql`（要求存在 `seed/`）。
- 各自目录内递归收集 `*.up.sql`，多目录时按文件路径字典序合并排序。
- `down` / `seed down` 各只回滚对应集合里的最后一步（`RollbackLast`）。

## 命名规范

统一使用：

- `<timestamp>_<action>.up.sql`
- `<timestamp>_<action>.down.sql`

示例：

- `202501011210_create_rbac.up.sql`
- `202501011210_create_rbac.down.sql`

建议：

- `schema/` 与 `seed/` 都使用同一套时间戳轴，保证全局顺序可控。
- seed 文件时间戳应晚于依赖它的 schema 文件。

## 编写约束

- `schema/*.up.sql` 只做 DDL，不写 seed 数据 `INSERT`。
- `seed/*.up.sql` 只做 DML，不做破坏式结构操作。
- `down.sql` 应与对应 `up.sql` 语义对称，且尽量幂等。
- 推荐使用 `IF EXISTS` / `IF NOT EXISTS` / `INSERT IGNORE` 等降低重复执行风险。

## 回滚注意事项

- `down` 是“最后一步回滚”，不是自动回滚到任意版本。
- 若某个 seed 依赖其他数据，回滚时要先删关联再删主数据。
- 线上回滚前先在影子库验证，避免因数据漂移导致失败。

## 协作建议

- 新增表或字段：优先放 `schema/`。
- 新增默认角色、菜单、账号：放 `seed/`。
- 不要修改已发布 migration 的历史内容，追加新 migration 修正。
- 每次提交 migration 后，至少本地验证一次：
  - `go run ./cmd/migrate up --env dev`
  - `go run ./cmd/migrate seed up --env dev`
  - `go run ./cmd/migrate down --env dev` / `go run ./cmd/migrate seed down --env dev`
