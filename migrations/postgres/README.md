# PostgreSQL Migrations Guide

本目录为 **PostgreSQL** 专用迁移 SQL，与 `migrations/mysql/` **同名时间戳** 对齐，便于双驱动对照维护。

## 环境要求

- 建议 **PostgreSQL 11+**（脚本使用 `ADD COLUMN IF NOT EXISTS` 等语法）。

## 目录约定

与 MySQL 保持一致的职责拆分：

- `schema/`：仅 `*_create_<表名>.{up,down}.sql`（与 MySQL 同名时间戳对齐时便于对照）。
- `seed/`：仅 `*_seed_<表名>.{up,down}.sql`，每张表一对。

## 执行规则

- 迁移入口为 `cmd/migrate`，`--driver postgres` 时默认根目录为 `migrations/postgres`。
- 与 MySQL 相同：`up` / `down` 只处理 `schema/`（若不存在则扫描整个根目录）；`seed up` / `seed down` 只处理 `seed/`（目录需存在）。
- 各自递归收集 `*.up.sql`，按文件路径字典序排序；回滚各集合内最后一步（`RollbackLast`）。

## 命名规范

统一使用：

- `<timestamp>_<action>.up.sql`
- `<timestamp>_<action>.down.sql`

示例：

- `202501011200_create_users.up.sql`
- `202501011200_create_users.down.sql`

建议：

- `schema/` 与 `seed/` 使用同一套时间戳轴，保证全局顺序可控。
- seed 文件时间戳应晚于依赖它的 schema 文件。

## 编写约束

- `schema/*.up.sql` 只做 DDL，不写 seed 数据 `INSERT`。
- `seed/*.up.sql` 只做 DML，且**只操作文件名中的那张表**；`down` 只删除本文件 `up` 写入的数据。
- `down.sql` 应与对应 `up.sql` 语义对称，且尽量幂等。
- 建议使用 PostgreSQL 的幂等语法（如 `IF EXISTS`、`IF NOT EXISTS`、`ON CONFLICT DO NOTHING`）。

## 回滚注意事项

- `down` 是“最后一步回滚”，不是自动回滚到任意版本。
- 若某个 seed 依赖其他数据，回滚时要先删关联再删主数据。
- 线上回滚前先在影子库验证，避免因数据漂移导致失败。

## 协作建议

- 新增表或字段：优先放 `schema/`。
- 新增默认角色、菜单、账号：放 `seed/`。
- 不要修改已发布 migration 的历史内容，追加新 migration 修正。
- 每次提交 migration 后，至少本地验证一次：
  - `go run ./cmd/migrate up --env dev --driver postgres --dsn "<your_pg_dsn>"`
  - `go run ./cmd/migrate seed up --env dev --driver postgres --dsn "<your_pg_dsn>"`
  - `go run ./cmd/migrate down --env dev --driver postgres --dsn "<your_pg_dsn>"` / `seed down`

