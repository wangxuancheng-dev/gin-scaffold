package dao

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/tenant"
)

func TestUserDAO_GetPrimaryRoles_emptyInput(t *testing.T) {
	d := &UserDAO{}
	m, err := d.GetPrimaryRoles(context.Background(), []int64{})
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = d.GetPrimaryRoles(context.Background(), nil)
	require.NoError(t, err)
	require.Empty(t, m)
}

func TestNewUserDAO(t *testing.T) {
	db, _ := newTestDB(t)
	d := NewUserDAO(db)
	require.NotNil(t, d)
}

func TestUserDAO_Create(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewUserDAO(db)
	ctx := context.Background()
	u := &model.User{Username: "alice", Password: "h", Nickname: "A", TenantID: "t1"}
	mock.ExpectExec("INSERT INTO `users`").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, d.Create(ctx, u))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserDAO_Create_tenantFromContext(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewUserDAO(db)
	ctx := tenant.WithContext(context.Background(), "acme")
	u := &model.User{Username: "bob", Password: "h", Nickname: "B"}
	mock.ExpectExec("INSERT INTO `users`").WillReturnResult(sqlmock.NewResult(2, 1))
	require.NoError(t, d.Create(ctx, u))
	require.Equal(t, "acme", u.TenantID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserDAO_Create_defaultTenant(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewUserDAO(db)
	u := &model.User{Username: "carl", Password: "h", Nickname: "C"}
	mock.ExpectExec("INSERT INTO `users`").WillReturnResult(sqlmock.NewResult(3, 1))
	require.NoError(t, d.Create(context.Background(), u))
	require.Equal(t, "default", u.TenantID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserDAO_CreateTx_nilTx(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewUserDAO(db)
	u := &model.User{Username: "d", Password: "p", Nickname: "n", TenantID: "t"}
	mock.ExpectExec("INSERT INTO `users`").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, d.CreateTx(context.Background(), nil, u))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserDAO_GetByID(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewUserDAO(db)
	ctx := tenant.WithContext(context.Background(), "t1")
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "username", "password", "nickname", "created_at", "updated_at", "deleted_at"}).
		AddRow(int64(1), "t1", "alice", "h", "A", now, now, nil)
	mock.ExpectQuery("SELECT .* FROM `users`").WillReturnRows(rows)
	u, err := d.GetByID(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, "alice", u.Username)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserDAO_GetByUsername(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewUserDAO(db)
	ctx := tenant.WithContext(context.Background(), "t1")
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "username", "password", "nickname", "created_at", "updated_at", "deleted_at"}).
		AddRow(int64(2), "t1", "bob", "x", "B", now, now, nil)
	mock.ExpectQuery("SELECT .* FROM `users`").WillReturnRows(rows)
	u, err := d.GetByUsername(ctx, "bob")
	require.NoError(t, err)
	require.Equal(t, int64(2), u.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserDAO_List(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewUserDAO(db)
	ctx := tenant.WithContext(context.Background(), "t1")
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(int64(1))
	dataRows := sqlmock.NewRows([]string{"id", "tenant_id", "username", "password", "nickname", "created_at", "updated_at", "deleted_at"}).
		AddRow(int64(1), "t1", "u", "p", "n", now, now, nil)
	mock.ExpectQuery("SELECT count\\(\\*\\)").WillReturnRows(countRows)
	mock.ExpectQuery("SELECT .* FROM `users`").WillReturnRows(dataRows)
	rows, total, err := d.List(ctx, model.UserQuery{}, 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, rows, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserDAO_ListForExport(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewUserDAO(db)
	ctx := tenant.WithContext(context.Background(), "t1")
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dataRows := sqlmock.NewRows([]string{"id", "tenant_id", "username", "password", "nickname", "created_at", "updated_at", "deleted_at"}).
		AddRow(int64(1), "t1", "u", "p", "n", now, now, nil)
	mock.ExpectQuery("SELECT .* FROM `users`").WillReturnRows(dataRows)
	rows, err := d.ListForExport(ctx, model.UserQuery{}, 100)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserDAO_ListAfterID(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewUserDAO(db)
	ctx := tenant.WithContext(context.Background(), "t1")
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dataRows := sqlmock.NewRows([]string{"id", "tenant_id", "username", "password", "nickname", "created_at", "updated_at", "deleted_at"}).
		AddRow(int64(5), "t1", "u", "p", "n", now, now, nil)
	mock.ExpectQuery("SELECT .* FROM `users`").WillReturnRows(dataRows)
	rows, err := d.ListAfterID(ctx, model.UserQuery{}, 3, 50)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}
