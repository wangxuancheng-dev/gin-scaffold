package model

import (
	"time"

	"gorm.io/gorm"
)

// Menu 后台菜单表。
type Menu struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"size:128;not null" json:"name"`
	Path      string         `gorm:"size:255;not null" json:"path"`
	PermCode  string         `gorm:"size:128;not null" json:"perm_code"`
	Sort      int            `gorm:"not null;default:0" json:"sort"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Menu) TableName() string {
	return "menus"
}
