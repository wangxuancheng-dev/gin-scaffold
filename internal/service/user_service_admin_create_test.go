package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"gin-scaffold/internal/config"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
)

func testJWTManager(t *testing.T) *jwtpkg.Manager {
	t.Helper()
	return jwtpkg.NewManager(&config.JWTConfig{
		Secret:           "unit-test-secret-for-service-tests",
		AccessExpireMin:  30,
		RefreshExpireMin: 60,
		Issuer:           "unit-test",
	})
}

func TestUserService_AdminCreate_usernameExists(t *testing.T) {
	repo := new(userServiceRepoMock)
	jm := testJWTManager(t)
	svc := NewUserService(repo, nil, jm, 0, nil, nil, config.OutboxConfig{})

	repo.On("GetByUsername", mock.Anything, "taken").Return(&model.User{ID: 1, Username: "taken"}, nil).Once()

	_, err := svc.AdminCreate(context.Background(), "taken", "password123", "nick", "user")
	require.Error(t, err)
	var biz *errcode.BizError
	require.ErrorAs(t, err, &biz)
	require.Equal(t, errcode.UserExists, biz.Code)
	repo.AssertExpectations(t)
}

func TestUserService_AdminCreate_getByUsernameNonNotFoundError(t *testing.T) {
	repo := new(userServiceRepoMock)
	jm := testJWTManager(t)
	svc := NewUserService(repo, nil, jm, 0, nil, nil, config.OutboxConfig{})

	repo.On("GetByUsername", mock.Anything, "x").Return((*model.User)(nil), errors.New("db unavailable")).Once()

	_, err := svc.AdminCreate(context.Background(), "x", "password123", "nick", "user")
	require.Error(t, err)
	require.Equal(t, "db unavailable", err.Error())
	repo.AssertExpectations(t)
}

func TestUserService_AdminCreate_defaultRole(t *testing.T) {
	repo := new(userServiceRepoMock)
	jm := testJWTManager(t)
	svc := NewUserService(repo, nil, jm, 0, nil, nil, config.OutboxConfig{})

	repo.On("GetByUsername", mock.Anything, "newuser").Return((*model.User)(nil), gorm.ErrRecordNotFound).Once()
	repo.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
		return u.Username == "newuser" && u.Nickname == "N"
	})).Return(nil).Once()
	repo.On("SetRole", mock.Anything, int64(0), "user").Return(nil).Once()

	u, err := svc.AdminCreate(context.Background(), "newuser", "password123", "N", "")
	require.NoError(t, err)
	require.NotNil(t, u)
	require.Equal(t, "newuser", u.Username)
	repo.AssertExpectations(t)
}
