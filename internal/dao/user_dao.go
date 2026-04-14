// Package dao 封装单表数据访问（无业务逻辑）。
package dao

import (
	"context"

	"gorm.io/gorm"

	"gin-scaffold/internal/model"
)

// UserDAO 用户表 DAO。
type UserDAO struct {
	db *gorm.DB
}

// NewUserDAO 构造。
func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

// Create 插入用户。
func (d *UserDAO) Create(ctx context.Context, u *model.User) error {
	return d.db.WithContext(ctx).Create(u).Error
}

// GetByID 主键查询。
func (d *UserDAO) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	if err := d.db.WithContext(ctx).First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByUsername 按用户名查询。
func (d *UserDAO) GetByUsername(ctx context.Context, name string) (*model.User, error) {
	var u model.User
	if err := d.db.WithContext(ctx).Where("username = ?", name).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// List 分页列表。
func (d *UserDAO) List(ctx context.Context, offset, limit int) ([]model.User, int64, error) {
	var total int64
	if err := d.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.User
	if err := d.db.WithContext(ctx).Order("id desc").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
