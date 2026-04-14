// Package dao 封装单表数据访问（无业务逻辑）。
package dao

import (
	"context"
	"errors"
	"time"

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

// GetByUsernameWithDeleted 按用户名查询（包含软删除）。
func (d *UserDAO) GetByUsernameWithDeleted(ctx context.Context, name string) (*model.User, error) {
	var u model.User
	if err := d.db.WithContext(ctx).Unscoped().Where("username = ?", name).First(&u).Error; err != nil {
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

// BindRole 绑定用户角色（若已存在则跳过）。
func (d *UserDAO) BindRole(ctx context.Context, userID int64, role string) error {
	var row struct {
		ID        int64
		DeletedAt *time.Time
	}
	err := d.db.WithContext(ctx).
		Table("user_roles").
		Select("id, deleted_at").
		Where("user_id = ? AND role = ?", userID, role).
		Order("id ASC").
		Limit(1).
		Take(&row).Error
	if err == nil {
		if row.DeletedAt == nil {
			return nil
		}
		return d.db.WithContext(ctx).Table("user_roles").Where("id = ?", row.ID).Updates(map[string]interface{}{
			"deleted_at": nil,
			"updated_at": time.Now(),
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return d.db.WithContext(ctx).Table("user_roles").Create(map[string]interface{}{
		"user_id": userID,
		"role":    role,
	}).Error
}

// Restore 通过主键恢复软删除用户并更新关键字段。
func (d *UserDAO) Restore(ctx context.Context, id int64, hashedPassword, nickname string) (*model.User, error) {
	if err := d.db.WithContext(ctx).Unscoped().Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"password":   hashedPassword,
		"nickname":   nickname,
		"deleted_at": nil,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return nil, err
	}
	var u model.User
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// GetPrimaryRole 查询用户主角色（按 user_roles.id 最早一条）。
func (d *UserDAO) GetPrimaryRole(ctx context.Context, userID int64) (string, error) {
	var row struct {
		Role string
	}
	err := d.db.WithContext(ctx).
		Table("user_roles").
		Select("role").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("id ASC").
		Limit(1).
		Take(&row).Error
	if err != nil {
		return "", err
	}
	return row.Role, nil
}
