package adminhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	adminreq "gin-scaffold/api/request/admin"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/internal/service/port"
)

type stubSystemSettingService struct {
	listRows []model.SystemSetting
	listTot  int64
	listErr  error

	getRow *model.SystemSetting
	getErr error

	createRow *model.SystemSetting
	createErr error
}

func (s *stubSystemSettingService) List(ctx context.Context, q model.SystemSettingQuery, page, pageSize int) ([]model.SystemSetting, int64, error) {
	return s.listRows, s.listTot, s.listErr
}

func (s *stubSystemSettingService) GetByID(ctx context.Context, id int64) (*model.SystemSetting, error) {
	return s.getRow, s.getErr
}

func (s *stubSystemSettingService) Create(ctx context.Context, key, value, valueType, groupName, remark string, actor model.SettingActor) (*model.SystemSetting, error) {
	return s.createRow, s.createErr
}

func (s *stubSystemSettingService) Update(ctx context.Context, id int64, value, valueType, groupName, remark *string, actor model.SettingActor) (*model.SystemSetting, error) {
	return nil, errors.New("not used")
}

func (s *stubSystemSettingService) Delete(ctx context.Context, id int64, actor model.SettingActor) error {
	return errors.New("not used")
}

func (s *stubSystemSettingService) Publish(ctx context.Context, id int64, note string, actor model.SettingActor) (*model.SystemSetting, error) {
	return nil, errors.New("not used")
}

func (s *stubSystemSettingService) ListHistory(ctx context.Context, id int64, page, pageSize int) ([]model.SystemSettingHistory, int64, error) {
	return nil, 0, errors.New("not used")
}

func (s *stubSystemSettingService) Rollback(ctx context.Context, id int64, historyID int64, reason string, actor model.SettingActor) (*model.SystemSetting, error) {
	return nil, errors.New("not used")
}

var _ port.SystemSettingService = (*stubSystemSettingService)(nil)

func TestSystemSettingHandler_List_ok(t *testing.T) {
	svc := &stubSystemSettingService{
		listRows: []model.SystemSetting{{ID: 1, Key: "k1"}},
		listTot:  1,
	}
	h := NewSystemSettingHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/system-settings?page=1&page_size=5", nil)
	h.List(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestSystemSettingHandler_Get_notFound(t *testing.T) {
	svc := &stubSystemSettingService{getErr: errcode.New(errcode.NotFound, errcode.KeyNotFound)}
	h := NewSystemSettingHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/system-settings/9", nil)
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	h.Get(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestSystemSettingHandler_Create_ok(t *testing.T) {
	svc := &stubSystemSettingService{
		createRow: &model.SystemSetting{ID: 2, Key: "site.name", Value: "x"},
	}
	h := NewSystemSettingHandler(svc)
	body, _ := json.Marshal(adminreq.SystemSettingCreateRequest{
		Key: "site.name", Value: "x", ValueType: "string", GroupName: "g", Remark: "r",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(jwtClaimsKey, &jwtpkg.Claims{UserID: 1, Role: "admin", TenantID: "default"})
	c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/admin/system-settings", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Create(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}
