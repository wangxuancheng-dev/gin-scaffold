# 数据库与 GORM 实践

## 推荐阅读顺序

1. 先看 **[配置说明](/guide/configuration)**（确认 `db.*`、`TIME_ZONE`、连接池参数）
2. 再看 **[数据库迁移与填充](/guide/database-and-migrations)**（确认表结构与种子数据）
3. 最后回到本页按“DAO 模板 -> 事务与锁 -> 测试策略”落地

## 相关文档回链

- 迁移与回滚规范：**[数据库迁移与填充](/guide/database-and-migrations)**
- 配置与环境变量覆盖：**[配置说明](/guide/configuration)**、**[环境变量绑定一览](/guide/environment-variables)**
- 日志与慢查询排障：**[日志与可观测](/guide/logging-observability)**
- 测试分层与执行方式：**[测试指南](/guide/testing-guide)**

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

## 常见 ORM 操作（DAO 模板）

以下示例建议都放在 DAO 层，Service 只做编排。

### 新增（Create）

```go
func (d *WidgetDAO) Create(ctx context.Context, in *model.Widget) error {
    return d.db.WithContext(ctx).Create(in).Error
}
```

### 按 ID 查询（First）

```go
func (d *WidgetDAO) GetByID(ctx context.Context, id int64) (*model.Widget, error) {
    q := tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id")
    var row model.Widget
    if err := q.Where("id = ?", id).First(&row).Error; err != nil {
        return nil, err
    }
    return &row, nil
}
```

### 列表 + 分页（Find + Count）

```go
func (d *WidgetDAO) List(ctx context.Context, page, pageSize int) ([]model.Widget, int64, error) {
    if page < 1 {
        page = 1
    }
    if pageSize <= 0 || pageSize > 100 {
        pageSize = 20
    }
    q := tenant.ApplyScope(ctx, d.db.WithContext(ctx).Model(&model.Widget{}), "tenant_id")
    var total int64
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    var list []model.Widget
    if err := q.Order("id desc").Offset((page-1)*pageSize).Limit(pageSize).Find(&list).Error; err != nil {
        return nil, 0, err
    }
    return list, total, nil
}
```

### 更新（Updates）

```go
func (d *WidgetDAO) UpdateName(ctx context.Context, id int64, name string) error {
    q := tenant.ApplyScope(ctx, d.db.WithContext(ctx).Model(&model.Widget{}), "tenant_id")
    return q.Where("id = ?", id).Updates(map[string]any{"name": name}).Error
}
```

### 删除（软删 / 硬删）

```go
func (d *WidgetDAO) SoftDelete(ctx context.Context, id int64) error {
    q := tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id")
    return q.Where("id = ?", id).Delete(&model.Widget{}).Error
}

func (d *WidgetDAO) HardDelete(ctx context.Context, id int64) error {
    q := tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id")
    return q.Unscoped().Where("id = ?", id).Delete(&model.Widget{}).Error
}
```

## 事务与锁（进阶）

### 行级锁（SELECT ... FOR UPDATE）

```go
import "gorm.io/gorm/clause"

func (d *WidgetDAO) LockByID(ctx context.Context, tx *gorm.DB, id int64) (*model.Widget, error) {
    q := tenant.ApplyScope(ctx, tx.WithContext(ctx), "tenant_id")
    var row model.Widget
    if err := q.Clauses(clause.Locking{Strength: "UPDATE"}).
        Where("id = ?", id).First(&row).Error; err != nil {
        return nil, err
    }
    return &row, nil
}
```

### 强一致读取（主库）

启用 `db.replicas` 后，如需“读己之写”，可在事务内查询或显式走主库（`dbresolver.Write`）。

## 关联查询与 N+1 避坑

### 预加载关联（Preload）

```go
func (d *OrderDAO) ListWithItems(ctx context.Context, page, pageSize int) ([]model.Order, error) {
    q := tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id")
    var rows []model.Order
    err := q.Preload("Items").
        Order("id desc").
        Offset((page-1)*pageSize).
        Limit(pageSize).
        Find(&rows).Error
    return rows, err
}
```

### 只加载必要字段（减少传输和扫描）

```go
func (d *OrderDAO) ListLite(ctx context.Context, page, pageSize int) ([]model.Order, error) {
    q := tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id")
    var rows []model.Order
    err := q.Select("id", "order_no", "status", "created_at").
        Order("id desc").
        Offset((page-1)*pageSize).
        Limit(pageSize).
        Find(&rows).Error
    return rows, err
}
```

建议：

- 列表页默认不要 `Preload` 过多关联，优先分层接口（摘要列表 + 详情接口）；
- 对高频查询先做 `Select` 字段收缩，再补索引；
- 排查 N+1 时先看慢 SQL 日志中是否出现大量重复 `SELECT ... WHERE id = ?`。

## 批量写入与幂等更新

### 批量插入（CreateInBatches）

```go
func (d *WidgetDAO) BatchCreate(ctx context.Context, rows []model.Widget) error {
    if len(rows) == 0 {
        return nil
    }
    return d.db.WithContext(ctx).CreateInBatches(rows, 200).Error
}
```

### Upsert（冲突更新）

```go
import "gorm.io/gorm/clause"

func (d *WidgetDAO) UpsertByCode(ctx context.Context, in *model.Widget) error {
    return d.db.WithContext(ctx).Clauses(clause.OnConflict{
        Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "code"}},
        DoUpdates: clause.AssignmentColumns([]string{"name", "updated_at"}),
    }).Create(in).Error
}
```

### 空更新保护（避免误把零值写回）

```go
func (d *WidgetDAO) Patch(ctx context.Context, id int64, patch map[string]any) error {
    if len(patch) == 0 {
        return nil
    }
    q := tenant.ApplyScope(ctx, d.db.WithContext(ctx).Model(&model.Widget{}), "tenant_id")
    return q.Where("id = ?", id).Updates(patch).Error
}
```

## 事务传播统一模式（TxDAO）

推荐在 DAO 提供 `WithTx(tx)` 或 `NewWithDB(tx)`，避免事务里误用全局 `db`。

### DAO 写法

```go
type WidgetDAO struct {
    db *gorm.DB
}

func NewWidgetDAO(db *gorm.DB) *WidgetDAO { return &WidgetDAO{db: db} }

func (d *WidgetDAO) WithTx(tx *gorm.DB) *WidgetDAO {
    if tx == nil {
        return d
    }
    return &WidgetDAO{db: tx}
}
```

### Service 事务编排

```go
func (s *WidgetService) CreateAndBind(ctx context.Context, in *model.Widget) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        dao := s.widgetDAO.WithTx(tx)
        if err := dao.Create(ctx, in); err != nil {
            return err
        }
        if err := dao.BindOwner(ctx, in.ID, in.OwnerID); err != nil {
            return err
        }
        return nil
    })
}
```

落地建议：

- 事务函数内部只使用 `dao.WithTx(tx)` 返回的实例；
- 不在事务块中混用 `s.db` 与 `tx`；
- 对外部副作用（消息、HTTP）优先走 Outbox，避免事务已提交但副作用失败造成不一致。

## ORM 测试策略（sqlmock vs integration）

推荐两层组合，而不是二选一。

### 单元层（快）：sqlmock

适合验证：

- SQL 形状是否符合预期（Where/Order/Limit/Tx 提交回滚）；
- 错误分支（`record not found`、唯一键冲突、死锁重试）；
- DAO 返回值与错误映射。

不适合验证：

- 不同数据库方言差异（MySQL/PostgreSQL）；
- 实际索引命中与执行计划；
- 时区、字符集、连接池行为。

### 集成层（真）：真实 DB

适合验证：

- 迁移后的真实表结构与约束；
- 事务隔离、锁行为、主从读写一致性策略；
- 与租户作用域、分页、排序组合后的真实结果。

建议：

- 每个核心 DAO 至少 1 个集成用例（创建 -> 查询 -> 更新 -> 删除）；
- 每个关键事务流程至少 1 个并发/锁相关用例；
- 统一通过 `tests/integration` 执行，并复用 CI 数据库容器。

## 最小测试模板（DAO）

### sqlmock 模板

```go
func TestWidgetDAO_GetByID(t *testing.T) {
    db, mock, _ := sqlmock.New()
    gdb, _ := gorm.Open(mysql.New(mysql.Config{Conn: db}), &gorm.Config{})
    dao := NewWidgetDAO(gdb)

    mock.ExpectQuery("SELECT .* FROM `widgets` WHERE id = \\?").
        WithArgs(1).
        WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "w1"))

    got, err := dao.GetByID(context.Background(), 1)
    require.NoError(t, err)
    require.Equal(t, int64(1), got.ID)
}
```

### integration 模板

```go
func TestWidgetDAO_CRUD_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("integration")
    }
    // 1) 准备测试 DB（迁移已完成）
    // 2) 调用 DAO Create/Get/Update/Delete
    // 3) 断言真实库结果
}
```

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
