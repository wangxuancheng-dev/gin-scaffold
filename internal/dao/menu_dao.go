package dao

import (
	"context"
	"time"

	"gorm.io/gorm"

	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/tenant"
)

// MenuDAO 菜单数据访问。
type MenuDAO struct {
	db *gorm.DB
}

func NewMenuDAO(db *gorm.DB) *MenuDAO {
	return &MenuDAO{db: db}
}

// ListByRole 查询角色可见菜单。
func (d *MenuDAO) ListByRole(ctx context.Context, role string) ([]model.Menu, error) {
	var menus []model.Menu
	tenantID := tenant.FromContext(ctx)
	if tenantID == "" {
		tenantID = "default"
	}
	err := d.db.WithContext(ctx).
		Table("menus AS m").
		Select("m.id, m.tenant_id, m.name, m.path, m.perm_code, m.sort, m.parent_id, m.created_at, m.updated_at").
		Joins("JOIN role_menus rm ON rm.menu_id = m.id AND rm.deleted_at IS NULL AND rm.tenant_id = m.tenant_id").
		Where("m.tenant_id = ? AND rm.tenant_id = ? AND rm.role = ? AND m.deleted_at IS NULL", tenantID, tenantID, role).
		Order("m.sort ASC, m.id ASC").
		Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}

// ListAllByTenant 返回当前租户下全部菜单（用于路由表 / 管理端维护）。
func (d *MenuDAO) ListAllByTenant(ctx context.Context) ([]model.Menu, error) {
	var menus []model.Menu
	tx := tenant.ApplyScope(ctx, d.db.WithContext(ctx).Model(&model.Menu{}), "tenant_id")
	if err := tx.Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}

// GetByID 主键查询（租户内）。
func (d *MenuDAO) GetByID(ctx context.Context, id int64) (*model.Menu, error) {
	var m model.Menu
	if err := tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id").First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// Create 创建菜单并默认绑定到 admin 角色（便于立即可见）。
func (d *MenuDAO) Create(ctx context.Context, m *model.Menu) error {
	if m.TenantID == "" {
		m.TenantID = tenant.FromContext(ctx)
		if m.TenantID == "" {
			m.TenantID = "default"
		}
	}
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(m).Error; err != nil {
			return err
		}
		return tx.Exec(
			`INSERT IGNORE INTO role_menus (tenant_id, role, menu_id, created_at, updated_at) VALUES (?, 'admin', ?, NOW(), NOW())`,
			m.TenantID, m.ID,
		).Error
	})
}

// Save 全量保存菜单（含 parent_id 置空等场景）。
func (d *MenuDAO) Save(ctx context.Context, m *model.Menu) error {
	if m == nil || m.ID == 0 {
		return gorm.ErrInvalidData
	}
	return tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id").Where("id = ?", m.ID).Save(m).Error
}

// SoftDelete 软删菜单及其子树，并同步软删角色关联。
func (d *MenuDAO) SoftDelete(ctx context.Context, id int64) error {
	tid := tenant.FromContext(ctx)
	if tid == "" {
		tid = "default"
	}
	now := time.Now()
	var all []model.Menu
	if err := tenant.ApplyScope(ctx, d.db.WithContext(ctx).Model(&model.Menu{}), "tenant_id").Order("sort ASC, id ASC").Find(&all).Error; err != nil {
		return err
	}
	desc := descendantMenuIDs(all, id)
	ids := append(desc, id)
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tenant.ApplyScope(ctx, tx.Model(&model.Menu{}), "tenant_id").Where("id IN ?", ids).Delete(&model.Menu{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Table("role_menus").
			Where("tenant_id = ? AND menu_id IN ? AND deleted_at IS NULL", tid, ids).
			Updates(map[string]interface{}{
				"deleted_at": now,
				"updated_at": now,
			}).Error
	})
}

func descendantMenuIDs(menus []model.Menu, root int64) []int64 {
	children := make(map[int64][]int64)
	for _, m := range menus {
		if m.ParentID == nil || *m.ParentID <= 0 {
			continue
		}
		p := *m.ParentID
		children[p] = append(children[p], m.ID)
	}
	var out []int64
	var walk func(int64)
	walk = func(n int64) {
		for _, c := range children[n] {
			out = append(out, c)
			walk(c)
		}
	}
	walk(root)
	return out
}
