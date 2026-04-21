package unit_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	adminhandler "gin-scaffold/internal/api/handler/admin"
	clienthandler "gin-scaffold/internal/api/handler/client"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) Register(ctx context.Context, username, password, nickname string) (*model.User, error) {
	args := m.Called(ctx, username, password, nickname)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *mockUserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	args := m.Called(ctx, id)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *mockUserService) Login(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}

func (m *mockUserService) LoginWithRefresh(ctx context.Context, username, password string) (string, string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *mockUserService) RefreshAccess(ctx context.Context, refreshToken string) (string, string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *mockUserService) List(ctx context.Context, q model.UserQuery, page, pageSize int) ([]model.User, int64, error) {
	args := m.Called(ctx, q, page, pageSize)
	rows, _ := args.Get(0).([]model.User)
	return rows, args.Get(1).(int64), args.Error(2)
}

func (m *mockUserService) StreamExport(
	ctx context.Context,
	q model.UserQuery,
	page, pageSize, limit, batchSize int,
	pageOnly, withRole bool,
	consume func(model.UserExportRow) error,
) error {
	args := m.Called(ctx, q, page, pageSize, limit, batchSize, pageOnly, withRole, consume)
	return args.Error(0)
}

func (m *mockUserService) AdminCreate(ctx context.Context, username, password, nickname, role string) (*model.User, error) {
	args := m.Called(ctx, username, password, nickname, role)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *mockUserService) AdminUpdate(ctx context.Context, id int64, nickname, password, role *string) (*model.User, error) {
	args := m.Called(ctx, id, nickname, password, role)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *mockUserService) AdminDelete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestUserHandler_List_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := new(mockUserService)
	h := adminhandler.NewUserHandler(svc)

	r := gin.New()
	r.GET("/users", h.List)

	svc.On("List", mock.Anything, mock.AnythingOfType("model.UserQuery"), 0, 0).Return([]model.User{
		{ID: 1, Username: "u1", Nickname: "N1"},
		{ID: 2, Username: "u2", Nickname: "N2"},
	}, int64(2), nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.EqualValues(t, 200, body["code"])

	data, ok := body["data"].(map[string]any)
	require.True(t, ok)
	require.EqualValues(t, 2, data["total"])

	list, ok := data["list"].([]any)
	require.True(t, ok)
	require.Len(t, list, 2)
	svc.AssertExpectations(t)
}

func TestUserHandler_Register_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := new(mockUserService)
	h := clienthandler.NewUserHandler(svc)

	r := gin.New()
	r.POST("/users", h.Register)

	svc.On("Register", mock.Anything, "alice", "123456", "Alice").Return(&model.User{
		ID:       100,
		Username: "alice",
		Nickname: "Alice",
	}, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"username":"alice","password":"123456","nickname":"Alice"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.EqualValues(t, 200, body["code"])
	data, ok := body["data"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "alice", data["username"])
	svc.AssertExpectations(t)
}

func TestUserHandler_Login_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := new(mockUserService)
	h := clienthandler.NewUserHandler(svc)

	r := gin.New()
	r.POST("/login", h.Login)

	svc.On("LoginWithRefresh", mock.Anything, "alice", "123456").Return("jwt-token", "refresh-token", nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"username":"alice","password":"123456"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.EqualValues(t, 200, body["code"])
	data, ok := body["data"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "jwt-token", data["access_token"])
	require.Equal(t, "refresh-token", data["refresh_token"])
	svc.AssertExpectations(t)
}

func TestUserHandler_Get_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := new(mockUserService)
	h := clienthandler.NewUserHandler(svc)

	r := gin.New()
	r.GET("/users/:id", h.Get)

	svc.On("GetByID", mock.Anything, int64(100)).Return(&model.User{
		ID:       100,
		Username: "alice",
		Nickname: "Alice",
	}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/users/100", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.EqualValues(t, 200, body["code"])
	data, ok := body["data"].(map[string]any)
	require.True(t, ok)
	require.EqualValues(t, 100, data["id"])
	require.Equal(t, "alice", data["username"])
	svc.AssertExpectations(t)
}

func TestUserHandler_Delete_SuperAdminProtected(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := new(mockUserService)
	h := adminhandler.NewUserHandler(svc)

	r := gin.New()
	r.DELETE("/users/:id", h.Delete)

	svc.On("AdminDelete", mock.Anything, int64(1)).Return(errcode.New(errcode.Forbidden, errcode.KeySuperAdminProtected)).Once()

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.EqualValues(t, errcode.Forbidden, body["code"])
	require.Equal(t, "super admin cannot be deleted", body["msg"])
	svc.AssertExpectations(t)
}
