package unit_test

import (
	"context"
	"testing"

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

func (m *mockUserRepo) List(ctx context.Context, offset, limit int) ([]model.User, int64, error) {
	args := m.Called(ctx, offset, limit)
	rows, _ := args.Get(0).([]model.User)
	return rows, args.Get(1).(int64), args.Error(2)
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
	svc := service.NewUserService(repo, nil, jm)

	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	require.NoError(t, err)

	repo.On("GetByUsername", mock.Anything, "alice").Return(&model.User{
		ID:       1,
		Username: "alice",
		Password: string(hash),
	}, nil).Once()

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
	svc := service.NewUserService(repo, nil, jm)

	repo.On("GetByUsername", mock.Anything, "nobody").Return((*model.User)(nil), gorm.ErrRecordNotFound).Once()

	token, err := svc.Login(context.Background(), "nobody", "123456")
	require.Error(t, err)
	require.Empty(t, token)
	repo.AssertExpectations(t)
}
