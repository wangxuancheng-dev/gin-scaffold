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

	adminreq "gin-scaffold/internal/api/request/admin"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/internal/service/port"
)

// stubUserService implements port.UserService for admin handler tests.
type stubUserService struct {
	listRows []model.User
	listTot  int64
	listErr  error

	getUser *model.User
	getErr  error

	adminCreateUser *model.User
	adminCreateErr  error

	adminUpdateUser *model.User
	adminUpdateErr  error

	adminDeleteErr error
}

func (s *stubUserService) Register(ctx context.Context, username, password, nickname string) (*model.User, error) {
	return nil, errors.New("not used")
}

func (s *stubUserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return s.getUser, s.getErr
}

func (s *stubUserService) Login(ctx context.Context, username, password string) (string, error) {
	return "", errors.New("not used")
}

func (s *stubUserService) LoginWithRefresh(ctx context.Context, username, password string) (string, string, error) {
	return "", "", errors.New("not used")
}

func (s *stubUserService) RefreshAccess(ctx context.Context, refreshToken string) (string, string, error) {
	return "", "", errors.New("not used")
}

func (s *stubUserService) List(ctx context.Context, q model.UserQuery, page, pageSize int) ([]model.User, int64, error) {
	return s.listRows, s.listTot, s.listErr
}

func (s *stubUserService) AdminCreate(ctx context.Context, username, password, nickname, role string) (*model.User, error) {
	return s.adminCreateUser, s.adminCreateErr
}

func (s *stubUserService) AdminUpdate(ctx context.Context, id int64, nickname, password, role *string) (*model.User, error) {
	return s.adminUpdateUser, s.adminUpdateErr
}

func (s *stubUserService) AdminDelete(ctx context.Context, id int64) error {
	return s.adminDeleteErr
}

func (s *stubUserService) StreamExport(
	ctx context.Context,
	q model.UserQuery,
	page, pageSize, limit, batchSize int,
	pageOnly, withRole bool,
	consume func(model.UserExportRow) error,
) error {
	return errors.New("not used")
}

var _ port.UserService = (*stubUserService)(nil)

func TestUserHandler_List_ok(t *testing.T) {
	svc := &stubUserService{
		listRows: []model.User{{ID: 1, Username: "a"}},
		listTot:  1,
	}
	h := NewUserHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/users?page=1&page_size=10", nil)
	h.List(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUserHandler_List_serviceError(t *testing.T) {
	svc := &stubUserService{listErr: errors.New("db")}
	h := NewUserHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/users", nil)
	h.List(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestUserHandler_Get_notFound(t *testing.T) {
	svc := &stubUserService{getErr: errcode.New(errcode.UserNotFound, errcode.KeyUserNotFound)}
	h := NewUserHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/users/9", nil)
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	h.Get(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestUserHandler_Get_ok(t *testing.T) {
	svc := &stubUserService{getUser: &model.User{ID: 9, Username: "bob"}}
	h := NewUserHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/users/9", nil)
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	h.Get(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestUserHandler_Create_flow(t *testing.T) {
	t.Run("bad_json", func(t *testing.T) {
		h := NewUserHandler(&stubUserService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/admin/users", bytes.NewBufferString("{"))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Create(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("code=%d", w.Code)
		}
	})
	t.Run("biz_error", func(t *testing.T) {
		svc := &stubUserService{adminCreateErr: errcode.New(errcode.UserExists, errcode.KeyUserExists)}
		h := NewUserHandler(svc)
		body, _ := json.Marshal(adminreq.UserCreateRequest{Username: "alice", Password: "secret12", Nickname: "A", Role: "user"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/admin/users", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Create(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
		}
	})
	t.Run("ok", func(t *testing.T) {
		svc := &stubUserService{adminCreateUser: &model.User{ID: 3, Username: "new"}}
		h := NewUserHandler(svc)
		body, _ := json.Marshal(adminreq.UserCreateRequest{Username: "alice", Password: "secret12", Nickname: "A", Role: "user"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/admin/users", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Create(c)
		if w.Code != http.StatusOK {
			t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
		}
	})
}

func TestUserHandler_Update_ok(t *testing.T) {
	svc := &stubUserService{adminUpdateUser: &model.User{ID: 2, Username: "u"}}
	h := NewUserHandler(svc)
	n := "nn"
	body, _ := json.Marshal(adminreq.UserUpdateRequest{Nickname: &n})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "http://localhost/admin/users/2", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	h.Update(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUserHandler_Delete_ok(t *testing.T) {
	svc := &stubUserService{}
	h := NewUserHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "http://localhost/admin/users/2", nil)
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	h.Delete(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestUserHandler_Delete_userNotFound(t *testing.T) {
	svc := &stubUserService{adminDeleteErr: errcode.New(errcode.UserNotFound, errcode.KeyUserNotFound)}
	h := NewUserHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "http://localhost/admin/users/99", nil)
	c.Params = gin.Params{{Key: "id", Value: "99"}}
	h.Delete(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestUserHandler_ExportTaskCreate_queueUnavailable(t *testing.T) {
	h := NewUserHandler(&stubUserService{}) // no job client
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/admin/users/export/tasks", nil)
	h.ExportTaskCreate(c)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUserHandler_ExportTaskStatus_emptyTaskID(t *testing.T) {
	h := NewUserHandler(&stubUserService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/users/export/tasks/", nil)
	c.Params = gin.Params{{Key: "task_id", Value: ""}}
	h.ExportTaskStatus(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", w.Code)
	}
}
