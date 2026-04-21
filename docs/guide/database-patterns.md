# 数据库与 GORM 实践

## 初始化与全局 DB

- `pkg/db.Init(&cfg.DB)` 在 bootstrap 中调用，返回 `*gorm.DB` 并注入 DAO / Service。
- 全局读取：`pkg/db.DB()`（慎用；推荐依赖注入传入的 `*gorm.DB`）。

## 驱动与时区

- 支持 **`mysql`** 与 **`postgres`**（`db.driver`）。
- `db.time_zone` 可被环境变量 **`TIME_ZONE`** 覆盖；与进程 `time.Local` 对齐逻辑见 bootstrap 注释。

## 只读副本（`db.replicas`）

- 当 `configs` 中 `db.replicas` 非空时，使用 GORM 插件 **`dbresolver`**，策略为 **`RandomPolicy`**（随机选择副本）。
- 写仍走主库；读默认走副本（GORM 解析策略）。业务若有「读己之写」强一致需求，需在事务或 `Clauses(dbresolver.Write)` 等层面显式控制（按需查阅 GORM dbresolver 文档）。

## 事务

项目内常见写法：

```go
err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    // 使用 tx 调用 DAO 或内联 SQL；任一步 return err 即回滚
    return nil
})
```

- **Service 层**编排多表：`internal/service/user_service.go` 等。
- **DAO 层**封装单聚合内多步：`internal/dao/menu_dao.go` 等。

原则：**事务边界放在真正需要原子性的那一层**，避免在 handler 里直接开事务。

## 多租户（`tenant.ApplyScope`）

- `internal/pkg/tenant`：`FromContext` / `WithContext` / **`ApplyScope(ctx, db, "tenant_id")`**。
- 当上下文存在 `tenant_id` 时，`ApplyScope` 会追加 `WHERE tenant_id = ?`；未启用或未设置租户时等价于原查询。
- **DAO 查询列表、详情、更新、删除**均应走 `ApplyScope`，避免串租户。
- 需要跨租户运维的极少数场景（如超管数据修复）需单独评审，慎用 `Unscoped()`（示例见 `user_dao` 中与删除/用户名相关的路径）。

DAO 内典型写法（省略结构体定义，与 `internal/dao/*_dao.go` 一致）：

```go
import (
    "context"

    "gin-scaffold/internal/model"
    "gin-scaffold/internal/pkg/tenant"
    "gorm.io/gorm"
)

func (d *WidgetDAO) Get(ctx context.Context, id int64) (*model.Widget, error) {
    q := tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id")
    var row model.Widget
    if err := q.Where("id = ?", id).First(&row).Error; err != nil {
        return nil, err
    }
    return &row, nil
}
```

中间件或上游 handler 已把租户写入 `context`（见 `middleware/tenant.go`）；测试里可用 **`tenant.WithContext(ctx, "default")`** 模拟。

## 慢 SQL 与日志

- `db.slow_threshold_ms`：超过阈值的 SQL 由 GORM logger 输出。
- `db.log_level`：`silent|error|warn|info`。

## 迁移与结构演进

- 表结构变更走 **`cmd/migrate`** 与 `migrations/` SQL，见 [数据库迁移与填充](/guide/database-and-migrations)。

## 最小可复制验证

```bash
# 1) 启动服务（包含 DB 初始化）
go run ./cmd/server server --env dev

# 2) 执行一轮结构 + 种子迁移（确保数据可用）
go run ./cmd/migrate up --env dev
go run ./cmd/migrate seed up --env dev

# 3) 访问一个依赖 DB 的管理接口（示例）
curl -sS "http://127.0.0.1:8080/api/v1/admin/users?page=1&page_size=10" \
  -H "Authorization: Bearer <admin-jwt>"
```

验证点：

- 服务启动日志中 DB 连接成功，且无 `validate config` 相关错误；
- 接口可正常返回分页数据（非 5xx）；
- 在启用租户场景下，仅返回当前租户数据（`tenant_id` 生效）。

## 常见问题与排查

- 启动即报 DB 连接失败：优先核对 `db.driver` 与 DSN 是否匹配（MySQL/PostgreSQL 不可混用）。
- 读写一致性异常：启用了 `db.replicas` 时，强一致读取需显式走主库或放在事务内。
- 多租户数据串读：DAO 查询/更新遗漏 `tenant.ApplyScope`，需统一补齐租户过滤。
- 慢查询明显增多：先检查索引与 SQL 条件，再结合 `db.slow_threshold_ms` 与日志定位热点语句。
- 回滚或发布异常：结构变更应通过迁移脚本治理，并先运行 [数据库迁移与填充](/guide/database-and-migrations) 与 [配置说明（关键组）](/guide/configuration) 的检查流程。
