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

## 慢 SQL 与日志

- `db.slow_threshold_ms`：超过阈值的 SQL 由 GORM logger 输出。
- `db.log_level`：`silent|error|warn|info`。

## 迁移与结构演进

- 表结构变更走 **`cmd/migrate`** 与 `migrations/` SQL，见 [数据库迁移与填充](/guide/database-and-migrations)。
