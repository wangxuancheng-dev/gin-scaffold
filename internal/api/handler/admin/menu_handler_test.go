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
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/internal/service/port"
)

const jwtClaimsKey = "jwt_claims"

type stubMenuService struct {
	listByRole    []model.Menu
	listByRoleErr error
	listAll       []model.Menu
	listAllErr    error
	getMenu       *model.Menu
	getErr        error
	createMenu    *model.Menu
	createErr     error
	updateMenu    *model.Menu
	updateErr     error
	deleteErr     error
}

func (s *stubMenuService) ListByRole(ctx context.Context, role string) ([]model.Menu, error) {
	return s.listByRole, s.listByRoleErr
}

func (s *stubMenuService) ListAllByTenant(ctx context.Context) ([]model.Menu, error) {
	return s.listAll, s.listAllErr
}

func (s *stubMenuService) GetByID(ctx context.Context, id int64) (*model.Menu, error) {
	return s.getMenu, s.getErr
}

func (s *stubMenuService) Create(ctx context.Context, name, path, permCode string, sort int, parentID *int64) (*model.Menu, error) {
	return s.createMenu, s.createErr
}

func (s *stubMenuService) Update(ctx context.Context, id int64, name, path, permCode *string, sort *int, parentID *int64) (*model.Menu, error) {
	return s.updateMenu, s.updateErr
}

func (s *stubMenuService) Delete(ctx context.Context, id int64) error {
	return s.deleteErr
}

var _ port.MenuService = (*stubMenuService)(nil)

func TestMenuHandler_ListMine_unauthorized(t *testing.T) {
	h := NewMenuHandler(&stubMenuService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/menus", nil)
	h.ListMine(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestMenuHandler_ListMine_ok(t *testing.T) {
	h := NewMenuHandler(&stubMenuService{listByRole: []model.Menu{{ID: 1, Name: "Home", Path: "/", PermCode: "p", Sort: 0}}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(jwtClaimsKey, &jwtpkg.Claims{UserID: 1, Role: "admin"})
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/menus", nil)
	h.ListMine(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestMenuHandler_Catalog_ok(t *testing.T) {
	h := NewMenuHandler(&stubMenuService{listAll: []model.Menu{{ID: 1, Name: "A", Path: "/a", PermCode: "a", Sort: 1}}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/menus/catalog", nil)
	h.Catalog(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestMenuHandler_Catalog_serviceError(t *testing.T) {
	h := NewMenuHandler(&stubMenuService{listAllErr: errors.New("db")})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/menus/catalog", nil)
	h.Catalog(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestMenuHandler_Get_notFound(t *testing.T) {
	h := NewMenuHandler(&stubMenuService{getErr: errcode.New(errcode.NotFound, errcode.KeyNotFound)})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/menus/5", nil)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	h.Get(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestMenuHandler_Create_ok(t *testing.T) {
	h := NewMenuHandler(&stubMenuService{createMenu: &model.Menu{ID: 10, Name: "X", Path: "/x", PermCode: "x", Sort: 1}})
	body, _ := json.Marshal(adminreq.MenuCreateRequest{Name: "X", Path: "/x", PermCode: "x", Sort: 1})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/admin/menus", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Create(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}
