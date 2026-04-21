package unit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	adminhandler "gin-scaffold/internal/api/handler/admin"
	"gin-scaffold/internal/config"
	"gin-scaffold/internal/middleware"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/tenant"
)

type mockSystemSettingService struct {
	mock.Mock
}

func (m *mockSystemSettingService) List(ctx context.Context, q model.SystemSettingQuery, page, pageSize int) ([]model.SystemSetting, int64, error) {
	args := m.Called(ctx, q, page, pageSize)
	rows, _ := args.Get(0).([]model.SystemSetting)
	return rows, args.Get(1).(int64), args.Error(2)
}

func (m *mockSystemSettingService) GetByID(ctx context.Context, id int64) (*model.SystemSetting, error) {
	args := m.Called(ctx, id)
	row, _ := args.Get(0).(*model.SystemSetting)
	return row, args.Error(1)
}

func (m *mockSystemSettingService) Create(ctx context.Context, key, value, valueType, groupName, remark string, actor model.SettingActor) (*model.SystemSetting, error) {
	args := m.Called(ctx, key, value, valueType, groupName, remark, actor)
	row, _ := args.Get(0).(*model.SystemSetting)
	return row, args.Error(1)
}

func (m *mockSystemSettingService) Update(ctx context.Context, id int64, value, valueType, groupName, remark *string, actor model.SettingActor) (*model.SystemSetting, error) {
	args := m.Called(ctx, id, value, valueType, groupName, remark, actor)
	row, _ := args.Get(0).(*model.SystemSetting)
	return row, args.Error(1)
}

func (m *mockSystemSettingService) Delete(ctx context.Context, id int64, actor model.SettingActor) error {
	args := m.Called(ctx, id, actor)
	return args.Error(0)
}

func (m *mockSystemSettingService) Publish(ctx context.Context, id int64, note string, actor model.SettingActor) (*model.SystemSetting, error) {
	args := m.Called(ctx, id, note, actor)
	row, _ := args.Get(0).(*model.SystemSetting)
	return row, args.Error(1)
}

func (m *mockSystemSettingService) ListHistory(ctx context.Context, id int64, page, pageSize int) ([]model.SystemSettingHistory, int64, error) {
	args := m.Called(ctx, id, page, pageSize)
	rows, _ := args.Get(0).([]model.SystemSettingHistory)
	return rows, args.Get(1).(int64), args.Error(2)
}

func (m *mockSystemSettingService) Rollback(ctx context.Context, id int64, historyID int64, reason string, actor model.SettingActor) (*model.SystemSetting, error) {
	args := m.Called(ctx, id, historyID, reason, actor)
	row, _ := args.Get(0).(*model.SystemSetting)
	return row, args.Error(1)
}

func TestSystemSettingHandler_Publish_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := new(mockSystemSettingService)
	h := adminhandler.NewSystemSettingHandler(svc)
	r := gin.New()
	r.Use(middleware.Tenant(&config.TenantConfig{Enabled: true, Header: "X-Tenant-ID", DefaultID: "default"}))
	r.POST("/system-settings/:id/publish", h.Publish)

	row := &model.SystemSetting{ID: 9, Key: "feature_x", Value: "on", IsPublished: true}
	svc.On("Publish", mock.Anything, int64(9), "release", mock.AnythingOfType("model.SettingActor")).
		Run(func(args mock.Arguments) {
			ctx, _ := args.Get(0).(context.Context)
			require.Equal(t, "tenant-a", tenant.FromContext(ctx))
			actor, _ := args.Get(3).(model.SettingActor)
			require.Equal(t, "tenant-a", actor.TenantID)
		}).
		Return(row, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/system-settings/9/publish", strings.NewReader(`{"note":"release"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-a")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestSystemSettingHandler_History_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSystemSettingService)
	h := adminhandler.NewSystemSettingHandler(svc)
	r := gin.New()
	r.GET("/system-settings/:id/history", h.History)

	now := time.Now()
	svc.On("ListHistory", mock.Anything, int64(5), 1, 10).Return([]model.SystemSettingHistory{
		{ID: 1, SettingID: 5, Action: "publish", CreatedAt: now},
	}, int64(1), nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/system-settings/5/history?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.EqualValues(t, 200, body["code"])
	svc.AssertExpectations(t)
}

func TestSystemSettingHandler_Rollback_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSystemSettingService)
	h := adminhandler.NewSystemSettingHandler(svc)
	r := gin.New()
	r.POST("/system-settings/:id/rollback", h.Rollback)

	row := &model.SystemSetting{ID: 5, Key: "feature_x", Value: "off", IsPublished: true}
	svc.On("Rollback", mock.Anything, int64(5), int64(10), "hotfix", mock.AnythingOfType("model.SettingActor")).Return(row, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/system-settings/5/rollback", strings.NewReader(`{"history_id":10,"reason":"hotfix"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}
