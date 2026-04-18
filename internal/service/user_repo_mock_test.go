package service

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"gin-scaffold/internal/model"
)

// userServiceRepoMock is a testify mock implementing UserRepo for service-layer tests.
type userServiceRepoMock struct {
	mock.Mock
}

func (m *userServiceRepoMock) Create(ctx context.Context, u *model.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *userServiceRepoMock) CreateTx(ctx context.Context, tx *gorm.DB, u *model.User) error {
	args := m.Called(ctx, tx, u)
	return args.Error(0)
}

func (m *userServiceRepoMock) BindRole(ctx context.Context, userID int64, role string) error {
	args := m.Called(ctx, userID, role)
	return args.Error(0)
}

func (m *userServiceRepoMock) BindRoleTx(ctx context.Context, tx *gorm.DB, userID int64, role string) error {
	args := m.Called(ctx, tx, userID, role)
	return args.Error(0)
}

func (m *userServiceRepoMock) Restore(ctx context.Context, id int64, hashedPassword, nickname string) (*model.User, error) {
	args := m.Called(ctx, id, hashedPassword, nickname)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *userServiceRepoMock) RestoreTx(ctx context.Context, tx *gorm.DB, id int64, hashedPassword, nickname string) (*model.User, error) {
	args := m.Called(ctx, tx, id, hashedPassword, nickname)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *userServiceRepoMock) GetByID(ctx context.Context, id int64) (*model.User, error) {
	args := m.Called(ctx, id)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *userServiceRepoMock) GetByUsername(ctx context.Context, name string) (*model.User, error) {
	args := m.Called(ctx, name)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *userServiceRepoMock) GetByUsernameWithDeleted(ctx context.Context, name string) (*model.User, error) {
	args := m.Called(ctx, name)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *userServiceRepoMock) List(ctx context.Context, q model.UserQuery, offset, limit int) ([]model.User, int64, error) {
	args := m.Called(ctx, q, offset, limit)
	rows, _ := args.Get(0).([]model.User)
	return rows, args.Get(1).(int64), args.Error(2)
}

func (m *userServiceRepoMock) ListAfterID(ctx context.Context, q model.UserQuery, lastID int64, limit int) ([]model.User, error) {
	args := m.Called(ctx, q, lastID, limit)
	rows, _ := args.Get(0).([]model.User)
	return rows, args.Error(1)
}

func (m *userServiceRepoMock) GetPrimaryRole(ctx context.Context, userID int64) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *userServiceRepoMock) GetPrimaryRoles(ctx context.Context, userIDs []int64) (map[int64]string, error) {
	args := m.Called(ctx, userIDs)
	rows, _ := args.Get(0).(map[int64]string)
	return rows, args.Error(1)
}

func (m *userServiceRepoMock) Update(ctx context.Context, id int64, nickname *string, hashedPassword *string) (*model.User, error) {
	args := m.Called(ctx, id, nickname, hashedPassword)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *userServiceRepoMock) SoftDelete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *userServiceRepoMock) SetRole(ctx context.Context, userID int64, role string) error {
	args := m.Called(ctx, userID, role)
	return args.Error(0)
}
