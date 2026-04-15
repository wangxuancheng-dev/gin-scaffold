package unit_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"gin-scaffold/config"
	"gin-scaffold/internal/model"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/internal/service"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(ctx context.Context, u *model.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *mockUserRepo) BindRole(ctx context.Context, userID int64, role string) error {
	args := m.Called(ctx, userID, role)
	return args.Error(0)
}

func (m *mockUserRepo) Restore(ctx context.Context, id int64, hashedPassword, nickname string) (*model.User, error) {
	args := m.Called(ctx, id, hashedPassword, nickname)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	args := m.Called(ctx, id)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, name string) (*model.User, error) {
	args := m.Called(ctx, name)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *mockUserRepo) GetByUsernameWithDeleted(ctx context.Context, name string) (*model.User, error) {
	args := m.Called(ctx, name)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *mockUserRepo) List(ctx context.Context, q model.UserQuery, offset, limit int) ([]model.User, int64, error) {
	args := m.Called(ctx, q, offset, limit)
	rows, _ := args.Get(0).([]model.User)
	return rows, args.Get(1).(int64), args.Error(2)
}

func (m *mockUserRepo) ListAfterID(ctx context.Context, q model.UserQuery, lastID int64, limit int) ([]model.User, error) {
	args := m.Called(ctx, q, lastID, limit)
	rows, _ := args.Get(0).([]model.User)
	return rows, args.Error(1)
}

func (m *mockUserRepo) GetPrimaryRole(ctx context.Context, userID int64) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *mockUserRepo) GetPrimaryRoles(ctx context.Context, userIDs []int64) (map[int64]string, error) {
	args := m.Called(ctx, userIDs)
	rows, _ := args.Get(0).(map[int64]string)
	return rows, args.Error(1)
}

func (m *mockUserRepo) Update(ctx context.Context, id int64, nickname *string, hashedPassword *string) (*model.User, error) {
	args := m.Called(ctx, id, nickname, hashedPassword)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *mockUserRepo) SoftDelete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepo) SetRole(ctx context.Context, userID int64, role string) error {
	args := m.Called(ctx, userID, role)
	return args.Error(0)
}

func TestUserServiceLogin_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	jm := jwtpkg.NewManager(&config.JWTConfig{
		Secret:           "unit-test-secret",
		AccessExpireMin:  30,
		RefreshExpireMin: 60,
		Issuer:           "unit-test",
	})
	svc := service.NewUserService(repo, nil, jm, 0)

	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	require.NoError(t, err)

	repo.On("GetByUsername", mock.Anything, "alice").Return(&model.User{
		ID:       1,
		Username: "alice",
		Password: string(hash),
	}, nil).Once()
	repo.On("GetPrimaryRole", mock.Anything, int64(1)).Return("user", nil).Once()

	token, err := svc.Login(context.Background(), "alice", "123456")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	repo.AssertExpectations(t)
}

func TestUserServiceLogin_UserNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	jm := jwtpkg.NewManager(&config.JWTConfig{
		Secret:           "unit-test-secret",
		AccessExpireMin:  30,
		RefreshExpireMin: 60,
		Issuer:           "unit-test",
	})
	svc := service.NewUserService(repo, nil, jm, 0)

	repo.On("GetByUsername", mock.Anything, "nobody").Return((*model.User)(nil), gorm.ErrRecordNotFound).Once()

	token, err := svc.Login(context.Background(), "nobody", "123456")
	require.Error(t, err)
	require.Empty(t, token)
	repo.AssertExpectations(t)
}

func TestUserServiceRegister_RestoreSoftDeleted(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	jm := jwtpkg.NewManager(&config.JWTConfig{
		Secret:           "unit-test-secret",
		AccessExpireMin:  30,
		RefreshExpireMin: 60,
		Issuer:           "unit-test",
	})
	svc := service.NewUserService(repo, nil, jm, 0)

	repo.On("GetByUsername", mock.Anything, "alice").Return((*model.User)(nil), gorm.ErrRecordNotFound).Once()
	repo.On("GetByUsernameWithDeleted", mock.Anything, "alice").Return(&model.User{
		ID:        9,
		Username:  "alice",
		DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
	}, nil).Once()
	repo.On("Restore", mock.Anything, int64(9), mock.AnythingOfType("string"), "Alice2").Return(&model.User{
		ID:       9,
		Username: "alice",
		Nickname: "Alice2",
	}, nil).Once()
	repo.On("BindRole", mock.Anything, int64(9), "user").Return(nil).Once()

	u, err := svc.Register(context.Background(), "alice", "123456", "Alice2")
	require.NoError(t, err)
	require.NotNil(t, u)
	require.Equal(t, int64(9), u.ID)
	require.Equal(t, "alice", u.Username)
	repo.AssertExpectations(t)
}
