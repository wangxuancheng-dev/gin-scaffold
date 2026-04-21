package clienthandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/internal/service/port"
)

// fullUserMock implements port.UserService for handler tests.
type fullUserMock struct {
	regUser *model.User
	regErr  error

	getUser *model.User
	getErr  error

	loginAccess  string
	loginRefresh string
	loginErr     error

	refAccess  string
	refRefresh string
	refErr     error
}

func (m *fullUserMock) Register(ctx context.Context, username, password, nickname string) (*model.User, error) {
	_ = ctx
	_ = username
	_ = password
	_ = nickname
	return m.regUser, m.regErr
}

func (m *fullUserMock) GetByID(ctx context.Context, id int64) (*model.User, error) {
	_ = ctx
	_ = id
	return m.getUser, m.getErr
}

func (m *fullUserMock) Login(ctx context.Context, username, password string) (string, error) {
	_ = ctx
	_ = username
	_ = password
	return "", errors.New("Login not stubbed")
}

func (m *fullUserMock) LoginWithRefresh(ctx context.Context, username, password string) (string, string, error) {
	_ = ctx
	_ = username
	_ = password
	return m.loginAccess, m.loginRefresh, m.loginErr
}

func (m *fullUserMock) RefreshAccess(ctx context.Context, refreshToken string) (string, string, error) {
	_ = ctx
	_ = refreshToken
	return m.refAccess, m.refRefresh, m.refErr
}

func (m *fullUserMock) List(ctx context.Context, q model.UserQuery, page, pageSize int) ([]model.User, int64, error) {
	_ = ctx
	_ = q
	_ = page
	_ = pageSize
	return nil, 0, errors.New("List not stubbed")
}

func (m *fullUserMock) AdminCreate(ctx context.Context, username, password, nickname, role string) (*model.User, error) {
	_ = ctx
	_ = username
	_ = password
	_ = nickname
	_ = role
	return nil, errors.New("AdminCreate not stubbed")
}

func (m *fullUserMock) AdminUpdate(ctx context.Context, id int64, nickname, password, role *string) (*model.User, error) {
	_ = ctx
	_ = id
	_ = nickname
	_ = password
	_ = role
	return nil, errors.New("AdminUpdate not stubbed")
}

func (m *fullUserMock) AdminDelete(ctx context.Context, id int64) error {
	_ = ctx
	_ = id
	return errors.New("AdminDelete not stubbed")
}

func (m *fullUserMock) StreamExport(
	ctx context.Context,
	q model.UserQuery,
	page, pageSize, limit, batchSize int,
	pageOnly, withRole bool,
	consume func(model.UserExportRow) error,
) error {
	_ = ctx
	_ = q
	_ = page
	_ = pageSize
	_ = limit
	_ = batchSize
	_ = pageOnly
	_ = withRole
	_ = consume
	return errors.New("StreamExport not stubbed")
}

var _ port.UserService = (*fullUserMock)(nil)

func TestUserHandler_Register_invalidJSON(t *testing.T) {
	h := NewUserHandler(&fullUserMock{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("{"))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Register(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUserHandler_Register_validation(t *testing.T) {
	h := NewUserHandler(&fullUserMock{})
	body := map[string]string{"username": "ab", "password": "short", "nickname": "n"}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Register(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestUserHandler_Register_serviceError(t *testing.T) {
	h := NewUserHandler(&fullUserMock{regErr: errcode.New(errcode.UserExists, errcode.KeyUserExists)})
	body := map[string]string{"username": "alice", "password": "secret12", "nickname": "A"}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Register(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestUserHandler_Register_ok(t *testing.T) {
	u := &model.User{ID: 7, Username: "alice", Nickname: "A", CreatedAt: time.Unix(1, 0).UTC(), UpdatedAt: time.Unix(2, 0).UTC()}
	h := NewUserHandler(&fullUserMock{regUser: u})
	body := map[string]string{"username": "alice", "password": "secret12", "nickname": "A"}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Register(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUserHandler_Get_invalidURI(t *testing.T) {
	h := NewUserHandler(&fullUserMock{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users/x", nil)
	c.Params = gin.Params{{Key: "id", Value: "x"}}
	h.Get(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestUserHandler_Get_notFound(t *testing.T) {
	h := NewUserHandler(&fullUserMock{getErr: errcode.New(errcode.UserNotFound, errcode.KeyUserNotFound)})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users/9", nil)
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	h.Get(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUserHandler_Get_ok(t *testing.T) {
	u := &model.User{ID: 9, Username: "bob", Nickname: "B"}
	h := NewUserHandler(&fullUserMock{getUser: u})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users/9", nil)
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	h.Get(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestUserHandler_Login_flow(t *testing.T) {
	t.Run("bad_json", func(t *testing.T) {
		h := NewUserHandler(&fullUserMock{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("not-json"))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Login(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("code=%d", w.Code)
		}
	})
	t.Run("unauthorized", func(t *testing.T) {
		h := NewUserHandler(&fullUserMock{loginErr: errcode.New(errcode.UserInvalidPwd, errcode.KeyUnauthorized)})
		body := map[string]string{"username": "alice", "password": "secret12"}
		b, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Login(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("code=%d", w.Code)
		}
	})
	t.Run("ok", func(t *testing.T) {
		h := NewUserHandler(&fullUserMock{loginAccess: "a", loginRefresh: "r"})
		body := map[string]string{"username": "alice", "password": "secret12"}
		b, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Login(c)
		if w.Code != http.StatusOK {
			t.Fatalf("code=%d", w.Code)
		}
	})
}

func TestUserHandler_Refresh_flow(t *testing.T) {
	t.Run("bad_json", func(t *testing.T) {
		h := NewUserHandler(&fullUserMock{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString("{"))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Refresh(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("code=%d", w.Code)
		}
	})
	t.Run("unauthorized", func(t *testing.T) {
		h := NewUserHandler(&fullUserMock{refErr: errcode.New(errcode.Unauthorized, errcode.KeyUnauthorized)})
		body := map[string]string{"refresh_token": "x"}
		b, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Refresh(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("code=%d", w.Code)
		}
	})
	t.Run("ok", func(t *testing.T) {
		h := NewUserHandler(&fullUserMock{refAccess: "na", refRefresh: "nr"})
		body := map[string]string{"refresh_token": "old"}
		b, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Refresh(c)
		if w.Code != http.StatusOK {
			t.Fatalf("code=%d", w.Code)
		}
	})
}
