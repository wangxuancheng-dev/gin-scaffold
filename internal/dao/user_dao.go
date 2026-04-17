// Package dao 封装单表数据访问（无业务逻辑）。
package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/tenant"
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
	if u != nil && u.TenantID == "" {
		u.TenantID = tenant.FromContext(ctx)
		if u.TenantID == "" {
			u.TenantID = "default"
		}
	}
	return d.db.WithContext(ctx).Create(u).Error
}

// GetByID 主键查询。
func (d *UserDAO) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	if err := tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id").First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByUsername 按用户名查询。
func (d *UserDAO) GetByUsername(ctx context.Context, name string) (*model.User, error) {
	var u model.User
	if err := tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id").Where("username = ?", name).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByUsernameWithDeleted 按用户名查询（包含软删除）。
func (d *UserDAO) GetByUsernameWithDeleted(ctx context.Context, name string) (*model.User, error) {
	var u model.User
	if err := tenant.ApplyScope(ctx, d.db.WithContext(ctx).Unscoped(), "tenant_id").Where("username = ?", name).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (d *UserDAO) applyFilters(tx *gorm.DB, q model.UserQuery) *gorm.DB {
	if q.Username != "" {
		tx = tx.Where("username LIKE ?", "%"+q.Username+"%")
	}
	if q.Nickname != "" {
		tx = tx.Where("nickname LIKE ?", "%"+q.Nickname+"%")
	}
	return tx
}

// List 分页列表。
func (d *UserDAO) List(ctx context.Context, q model.UserQuery, offset, limit int) ([]model.User, int64, error) {
	tx := d.applyFilters(tenant.ApplyScope(ctx, d.db.WithContext(ctx).Model(&model.User{}), "tenant_id"), q)
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.User
	if err := d.applyFilters(tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id"), q).Order("id desc").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// ListForExport 导出列表（不返回 total）。
func (d *UserDAO) ListForExport(ctx context.Context, q model.UserQuery, limit int) ([]model.User, error) {
	var rows []model.User
	if err := d.applyFilters(tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id"), q).Order("id desc").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// ListAfterID 按主键增量扫描，适合大数据导出。
func (d *UserDAO) ListAfterID(ctx context.Context, q model.UserQuery, lastID int64, limit int) ([]model.User, error) {
	tx := d.applyFilters(tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id"), q)
	if lastID > 0 {
		tx = tx.Where("id > ?", lastID)
	}
	var rows []model.User
	if err := tx.Order("id asc").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// BindRole 绑定用户角色（若已存在则跳过）。
func (d *UserDAO) BindRole(ctx context.Context, userID int64, role string) error {
	tenantID := tenant.FromContext(ctx)
	if tenantID == "" {
		tenantID = "default"
	}
	var row struct {
		ID        int64
		DeletedAt *time.Time
	}
	err := d.db.WithContext(ctx).
		Table("user_roles").
		Select("id, deleted_at").
		Where("tenant_id = ? AND user_id = ? AND role = ?", tenantID, userID, role).
		Order("id ASC").
		Limit(1).
		Take(&row).Error
	if err == nil {
		if row.DeletedAt == nil {
			return nil
		}
		return d.db.WithContext(ctx).Table("user_roles").Where("tenant_id = ? AND id = ?", tenantID, row.ID).Updates(map[string]interface{}{
			"deleted_at": nil,
			"updated_at": time.Now(),
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return d.db.WithContext(ctx).Table("user_roles").Create(map[string]interface{}{
		"tenant_id": tenantID,
		"user_id":   userID,
		"role":      role,
	}).Error
}

// Restore 通过主键恢复软删除用户并更新关键字段。
func (d *UserDAO) Restore(ctx context.Context, id int64, hashedPassword, nickname string) (*model.User, error) {
	if err := tenant.ApplyScope(ctx, d.db.WithContext(ctx).Unscoped().Model(&model.User{}), "tenant_id").Where("id = ?", id).Updates(map[string]interface{}{
		"password":   hashedPassword,
		"nickname":   nickname,
		"deleted_at": nil,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return nil, err
	}
	var u model.User
	if err := tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id").Where("id = ?", id).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// Update 更新用户资料（支持昵称/密码）。
func (d *UserDAO) Update(ctx context.Context, id int64, nickname *string, hashedPassword *string) (*model.User, error) {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if nickname != nil {
		updates["nickname"] = *nickname
	}
	if hashedPassword != nil && *hashedPassword != "" {
		updates["password"] = *hashedPassword
	}
	if err := tenant.ApplyScope(ctx, d.db.WithContext(ctx).Model(&model.User{}), "tenant_id").Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return d.GetByID(ctx, id)
}

// SoftDelete 软删除用户。
func (d *UserDAO) SoftDelete(ctx context.Context, id int64) error {
	return tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id").Delete(&model.User{}, id).Error
}

// SetRole 将用户角色切换为给定角色（软删历史角色，保留审计）。
func (d *UserDAO) SetRole(ctx context.Context, userID int64, role string) error {
	tenantID := tenant.FromContext(ctx)
	if tenantID == "" {
		tenantID = "default"
	}
	if err := d.db.WithContext(ctx).Table("user_roles").
		Where("tenant_id = ? AND user_id = ? AND deleted_at IS NULL", tenantID, userID).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		}).Error; err != nil {
		return err
	}
	return d.BindRole(ctx, userID, role)
}

// GetPrimaryRole 查询用户主角色（按 user_roles.id 最早一条）。
func (d *UserDAO) GetPrimaryRole(ctx context.Context, userID int64) (string, error) {
	tenantID := tenant.FromContext(ctx)
	if tenantID == "" {
		tenantID = "default"
	}
	var row struct {
		Role string
	}
	err := d.db.WithContext(ctx).
		Table("user_roles").
		Select("role").
		Where("tenant_id = ? AND user_id = ? AND deleted_at IS NULL", tenantID, userID).
		Order("id ASC").
		Limit(1).
		Take(&row).Error
	if err != nil {
		return "", err
	}
	return row.Role, nil
}

// GetPrimaryRoles 批量查询用户主角色（每用户 user_roles.id 最早一条）。
func (d *UserDAO) GetPrimaryRoles(ctx context.Context, userIDs []int64) (map[int64]string, error) {
	out := make(map[int64]string, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}
	tenantID := tenant.FromContext(ctx)
	if tenantID == "" {
		tenantID = "default"
	}
	sub := d.db.WithContext(ctx).
		Table("user_roles").
		Select("user_id, MIN(id) AS min_id").
		Where("tenant_id = ? AND deleted_at IS NULL AND user_id IN ?", tenantID, userIDs).
		Group("user_id")

	var rows []struct {
		UserID int64
		Role   string
	}
	if err := d.db.WithContext(ctx).
		Table("user_roles ur").
		Select("ur.user_id, ur.role").
		Joins("JOIN (?) t ON ur.user_id = t.user_id AND ur.id = t.min_id", sub).
		Where("ur.tenant_id = ?", tenantID).
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("batch get roles: %w", err)
	}
	for _, r := range rows {
		out[r.UserID] = r.Role
	}
	return out, nil
}
