package dao

import (
	"context"

	"gorm.io/gorm"

	"gin-scaffold/internal/pkg/tenant"
)

// AuthzDAO 权限数据访问。
type AuthzDAO struct {
	db *gorm.DB
}

// NewAuthzDAO 构造。
func NewAuthzDAO(db *gorm.DB) *AuthzDAO {
	return &AuthzDAO{db: db}
}

// HasRolePermission 判断角色是否具备某权限。
func (d *AuthzDAO) HasRolePermission(ctx context.Context, role, permission string) (bool, error) {
	var count int64
	tenantID := tenant.FromContext(ctx)
	if tenantID == "" {
		tenantID = "default"
	}
	err := d.db.WithContext(ctx).
		Table("role_permissions").
		Where("tenant_id = ? AND role = ? AND permission = ? AND deleted_at IS NULL", tenantID, role, permission).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
