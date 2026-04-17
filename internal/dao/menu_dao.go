package dao

import (
	"context"

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
		Select("m.id, m.tenant_id, m.name, m.path, m.perm_code, m.sort, m.created_at, m.updated_at").
		Joins("JOIN role_menus rm ON rm.menu_id = m.id AND rm.deleted_at IS NULL AND rm.tenant_id = m.tenant_id").
		Where("m.tenant_id = ? AND rm.tenant_id = ? AND rm.role = ? AND m.deleted_at IS NULL", tenantID, tenantID, role).
		Order("m.sort ASC, m.id ASC").
		Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}
