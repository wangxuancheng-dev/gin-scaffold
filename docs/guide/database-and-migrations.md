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

## 新增表结构（示例）

在 `migrations/mysql/schema/`（或 `migrations/postgres/schema/`）新增一对文件，时间戳全局递增，例如：

- `202604191200_create_widgets.up.sql`
- `202604191200_create_widgets.down.sql`

MySQL 示例：

```sql
-- up
CREATE TABLE IF NOT EXISTS `widgets` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
  `name` VARCHAR(128) NOT NULL,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_widgets_tenant` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

`down` 文件对称删除，例如：

```sql
DROP TABLE IF EXISTS `widgets`;
```

## 新增种子数据（示例）

在 `migrations/mysql/seed/` 增加 `*_seed_<表名>.up.sql`，只做 **DML**，常用 **`INSERT IGNORE`** 或 **`INSERT ... ON DUPLICATE`** 保持可重复执行（详见各驱动目录下 `README.md`）。

```sql
INSERT IGNORE INTO `role_permissions` (`tenant_id`, `role_id`, `permission_code`, `created_at`, `updated_at`)
SELECT 'default', r.id, 'widget:read', NOW(), NOW()
FROM `roles` r WHERE r.tenant_id = 'default' AND r.code = 'admin'
AND NOT EXISTS (
  SELECT 1 FROM `role_permissions` rp WHERE rp.tenant_id = 'default' AND rp.role_id = r.id AND rp.permission_code = 'widget:read'
);
```

## 测试库

- 集成测试脚本会创建 `scaffold_test` 并执行 migrate + seed + fixture，见 `tests/integration/README.md` 与 `scripts/integration.sh`。

## 最小可复制验证

```bash
# 结构迁移
go run ./cmd/migrate up --env dev

# 种子迁移
go run ./cmd/migrate seed up --env dev

# 可选：回滚一轮（谨慎）
go run ./cmd/migrate seed down --env dev
go run ./cmd/migrate down --env dev
```

## 常见问题与排查

- `up` 成功但登录失败：通常是漏执行 `seed up`。
- `driver` 不匹配：MySQL DSN 却传了 `--driver postgres` 会直接失败。
- 迁移脚本回滚不对称：运行 `bash ./scripts/check-migration-lint.sh .` 检查 up/down 与高风险语句。
- 线上锁表风险：避免在一个事务里堆叠多条 `ALTER TABLE`；必要时分批发布。
- 权限菜单不生效：确认执行了 seed 且当前租户数据正确（默认 `default`）。

## 下一步

- 运行期访问数据库、事务与多租户：**[数据库与 GORM 实践](/guide/database-patterns)**。
