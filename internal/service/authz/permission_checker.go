package authz

import "context"

// RolePermissionRepo 权限检查所需的数据访问抽象。
type RolePermissionRepo interface {
	HasRolePermission(ctx context.Context, role, permission string) (bool, error)
}

// DBPermissionChecker 基于数据库角色权限表的检查器。
type DBPermissionChecker struct {
	repo             RolePermissionRepo
	superAdminUserID int64
}

// NewDBPermissionChecker 构造数据库权限检查器。
func NewDBPermissionChecker(repo RolePermissionRepo, superAdminUserID int64) *DBPermissionChecker {
	return &DBPermissionChecker{repo: repo, superAdminUserID: superAdminUserID}
}

// HasPermission 优先按数据库策略判断权限。
func (c *DBPermissionChecker) HasPermission(ctx context.Context, userID int64, role, permission string) (bool, error) {
	if c != nil && c.superAdminUserID > 0 && userID == c.superAdminUserID {
		return true, nil
	}
	if c == nil || c.repo == nil {
		return false, nil
	}
	return c.repo.HasRolePermission(ctx, role, permission)
}
