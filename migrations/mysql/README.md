# MySQL Migrations Guide

本目录按职责拆分为两个子目录：

- `schema/`：结构变更（DDL），如 `CREATE TABLE`、`ALTER TABLE`、索引与约束。
- `seed/`：初始化与演示数据（DML），如默认角色、菜单、管理员账号。

## 执行规则

- 迁移入口为 `cmd/migrate`，默认扫描 `migrations/mysql`。
- 扫描方式为递归扫描子目录，收集全部 `*.up.sql`。
- 执行顺序按文件名全局字典序排序（不是按目录先后）。
- 回滚 `down` 仅回滚最后一步（`RollbackLast`）。

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
  - `go run ./cmd/migrate --env dev up`
  - `go run ./cmd/migrate --env dev down`
