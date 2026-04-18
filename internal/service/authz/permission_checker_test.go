package authz

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeRoleRepo struct {
	has bool
	err error
}

func (f *fakeRoleRepo) HasRolePermission(ctx context.Context, role, permission string) (bool, error) {
	_ = ctx
	_ = role
	_ = permission
	return f.has, f.err
}

func TestDBPermissionChecker_HasPermission_superAdminBypass(t *testing.T) {
	c := NewDBPermissionChecker(&fakeRoleRepo{has: false}, 100)
	ok, err := c.HasPermission(context.Background(), 100, "admin", "x.y")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestDBPermissionChecker_HasPermission_delegatesToRepo(t *testing.T) {
	c := NewDBPermissionChecker(&fakeRoleRepo{has: true}, 0)
	ok, err := c.HasPermission(context.Background(), 1, "r", "p")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestDBPermissionChecker_HasPermission_repoError(t *testing.T) {
	c := NewDBPermissionChecker(&fakeRoleRepo{err: errors.New("db")}, 0)
	_, err := c.HasPermission(context.Background(), 1, "r", "p")
	require.Error(t, err)
}

func TestDBPermissionChecker_HasPermission_nilChecker(t *testing.T) {
	var c *DBPermissionChecker
	ok, err := c.HasPermission(context.Background(), 1, "r", "p")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestDBPermissionChecker_HasPermission_nilRepo(t *testing.T) {
	c := &DBPermissionChecker{repo: nil, superAdminUserID: 0}
	ok, err := c.HasPermission(context.Background(), 1, "r", "p")
	require.NoError(t, err)
	require.False(t, ok)
}
