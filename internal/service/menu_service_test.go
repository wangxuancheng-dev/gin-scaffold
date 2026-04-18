package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
)

type menuRepoMock struct {
	getByIDRes *model.Menu
	getByIDErr error
	softDelErr error
	listAll    []model.Menu
	listAllErr error
}

func (m *menuRepoMock) ListByRole(ctx context.Context, role string) ([]model.Menu, error) {
	_, _ = ctx, role
	return nil, nil
}

func (m *menuRepoMock) ListAllByTenant(ctx context.Context) ([]model.Menu, error) {
	_ = ctx
	return m.listAll, m.listAllErr
}

func (m *menuRepoMock) GetByID(ctx context.Context, id int64) (*model.Menu, error) {
	_, _ = ctx, id
	return m.getByIDRes, m.getByIDErr
}

func (m *menuRepoMock) Create(ctx context.Context, menu *model.Menu) error {
	_, _ = ctx, menu
	return nil
}

func (m *menuRepoMock) Save(ctx context.Context, menu *model.Menu) error {
	_, _ = ctx, menu
	return nil
}

func (m *menuRepoMock) SoftDelete(ctx context.Context, id int64) error {
	_, _ = ctx, id
	return m.softDelErr
}

func TestMenuService_GetByID_notFound(t *testing.T) {
	svc := NewMenuService(&menuRepoMock{getByIDErr: gorm.ErrRecordNotFound})
	_, err := svc.GetByID(context.Background(), 404)
	var biz *errcode.BizError
	require.ErrorAs(t, err, &biz)
	require.Equal(t, errcode.NotFound, biz.Code)
}

func TestMenuService_Delete_notFound(t *testing.T) {
	svc := NewMenuService(&menuRepoMock{softDelErr: gorm.ErrRecordNotFound})
	err := svc.Delete(context.Background(), 1)
	var biz *errcode.BizError
	require.ErrorAs(t, err, &biz)
	require.Equal(t, errcode.NotFound, biz.Code)
}

func ptrI64(v int64) *int64 {
	p := v
	return &p
}

func TestMenuDescendantContains_directAndNested(t *testing.T) {
	menus := []model.Menu{
		{ID: 1},
		{ID: 2, ParentID: ptrI64(1)},
		{ID: 3, ParentID: ptrI64(2)},
	}
	require.True(t, menuDescendantContains(menus, 1, 1))
	require.True(t, menuDescendantContains(menus, 1, 3))
	require.False(t, menuDescendantContains(menus, 3, 1))
	require.False(t, menuDescendantContains(menus, 1, 99))
}

func TestMenuDescendantContains_siblingBranch(t *testing.T) {
	menus := []model.Menu{
		{ID: 1},
		{ID: 2, ParentID: ptrI64(1)},
		{ID: 3, ParentID: ptrI64(1)},
	}
	require.False(t, menuDescendantContains(menus, 2, 3))
}
