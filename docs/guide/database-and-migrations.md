# 数据库迁移与填充（Seed）

## 工具入口

```bash
go run ./cmd/migrate up --env dev
go run ./cmd/migrate seed up --env dev
```

生产建议使用编译后的 `./bin/migrate`（见 [上线清单](/checklist)）。

## 目录约定

- **MySQL**：`migrations/mysql/schema/`（结构）、`migrations/mysql/seed/`（种子）。
- **PostgreSQL**：`migrations/postgres/schema/`、`migrations/postgres/seed/`。
- 未指定 `--dir` 时，`cmd/migrate` 按 **`--driver`** 自动选择目录（详见 `cmd/migrate` 内说明与各 `README.md`）。

## 常见参数

- `--env`：`dev` | `test` | `prod`，决定加载哪套 `configs/app*.yaml`。
- `--driver`：`mysql` 或 `postgres`。
- `--dsn`：覆盖配置文件中的 DSN（生产常用环境变量注入）。
- `--time-zone`：可覆盖会话时区（与 `db.time_zone` 独立）。

## 结构迁移 vs Seed

- **`up` / `down`**：只跑 schema 目录，负责表结构。
- **`seed up` / `seed down`**：只跑 seed 目录，负责权限、菜单、初始管理员等 **可重复执行策略** 由你们 SQL 设计决定（通常用 upsert 或幂等脚本）。

## 与 GORM 的关系

- 运行时使用 **GORM** 访问数据库；迁移文件为 **SQL**（也可用同一仓库内其他工具链，但本脚手架以 SQL 为准）。

## 测试库

- 集成测试脚本会创建 `scaffold_test` 并执行 migrate + seed + fixture，见 `tests/integration/README.md` 与 `scripts/integration.sh`。

## 下一步

- 运行期访问数据库、事务与多租户：**[数据库与 GORM 实践](/guide/database-patterns)**。
