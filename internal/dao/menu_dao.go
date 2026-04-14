package dao

import (
	"context"

	"gorm.io/gorm"

	"gin-scaffold/internal/model"
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
	err := d.db.WithContext(ctx).
		Table("menus AS m").
		Select("m.id, m.name, m.path, m.perm_code, m.sort, m.created_at, m.updated_at").
		Joins("JOIN role_menus rm ON rm.menu_id = m.id AND rm.deleted_at IS NULL").
		Where("rm.role = ? AND m.deleted_at IS NULL", role).
		Order("m.sort ASC, m.id ASC").
		Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}
