// Package model 定义数据库表模型（仅结构体与 GORM 标签）。
package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表。
type User struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  string         `gorm:"size:64;not null;default:default;uniqueIndex:uk_users_tenant_username,priority:1;index:idx_users_tenant" json:"tenant_id"`
	Username  string         `gorm:"size:64;not null;uniqueIndex:uk_users_tenant_username,priority:2" json:"username"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Nickname  string         `gorm:"size:64" json:"nickname"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名。
func (User) TableName() string {
	return "users"
}
