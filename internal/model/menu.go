package model

import (
	"time"

	"gorm.io/gorm"
)

// Menu 后台菜单表。
type Menu struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  string         `gorm:"size:64;not null;default:default;uniqueIndex:uk_menus_tenant_path,priority:1;index:idx_menus_tenant" json:"tenant_id"`
	Name      string         `gorm:"size:128;not null" json:"name"`
	Path      string         `gorm:"size:255;not null;uniqueIndex:uk_menus_tenant_path,priority:2" json:"path"`
	PermCode  string         `gorm:"size:128;not null" json:"perm_code"`
	Sort      int            `gorm:"not null;default:0" json:"sort"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Menu) TableName() string {
	return "menus"
}
