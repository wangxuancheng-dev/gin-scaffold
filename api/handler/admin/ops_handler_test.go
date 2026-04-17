package adminhandler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"gin-scaffold/internal/dao"
	"gin-scaffold/internal/model"
)

func TestParseAuditQueryExportDefaultWindow(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs/export", nil)

	got, ok := parseAuditQuery(c, true)
	if !ok {
		t.Fatalf("expected ok=true, status=%d body=%s", w.Code, w.Body.String())
	}
	if got.From == nil || got.To == nil {
		t.Fatalf("expected from/to defaulted, got from=%v to=%v", got.From, got.To)
	}
	diff := got.To.Sub(*got.From)
	expect := time.Duration(fallbackAuditExportDefaultDays) * 24 * time.Hour
	if diff < expect-3*time.Second || diff > expect+3*time.Second {
		t.Fatalf("expected about 7d window, got %s", diff)
	}
}

func TestParseAuditQueryExportRangeTooLarge(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	from := time.Now().Add(-40 * 24 * time.Hour).Format(time.RFC3339)
	to := time.Now().Format(time.RFC3339)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs/export?from="+from+"&to="+to, nil)

	_, ok := parseAuditQuery(c, true)
	if ok {
		t.Fatalf("expected ok=false for >31d range")
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestParseAuditQueryInvalidFrom(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs/export?from=invalid", nil)

	_, ok := parseAuditQuery(c, true)
	if ok {
		t.Fatalf("expected ok=false when from invalid")
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

type fakeAuditStore struct {
	rows []model.AuditLog
}

func (f *fakeAuditStore) Create(context.Context, *model.AuditLog) error {
	return nil
}

func (f *fakeAuditStore) List(context.Context, dao.AuditLogListQuery) ([]model.AuditLog, int64, error) {
	return []model.AuditLog{}, 0, nil
}

func (f *fakeAuditStore) ListForExport(context.Context, dao.AuditLogListQuery, int) ([]model.AuditLog, error) {
	return f.rows, nil
}

func TestAuditLogsExport_Headers(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	h := NewOpsHandler(&fakeAuditStore{
		rows: []model.AuditLog{
			{ID: 1, Action: "POST", Path: "/api/v1/admin/users", Status: 200, CreatedAt: time.Now()},
		},
	})
	r := gin.New()
	r.GET("/api/v1/admin/audit-logs/export", h.AuditLogsExport)
	from := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
	to := time.Now().UTC().Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs/export?from="+url.QueryEscape(from)+"&to="+url.QueryEscape(to)+"&limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	if w.Header().Get("X-Export-Count") != "1" {
		t.Fatalf("expected X-Export-Count=1, got %q", w.Header().Get("X-Export-Count"))
	}
	window := w.Header().Get("X-Export-Window")
	if window == "" || len(window) < 10 {
		t.Fatalf("expected non-empty X-Export-Window, got %q", window)
	}
}

func TestAuditLogsExport_LimitInvalid(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	h := NewOpsHandler(&fakeAuditStore{
		rows: []model.AuditLog{
			{ID: 1, Action: "POST", Path: "/api/v1/admin/users", Status: 200, CreatedAt: time.Now()},
		},
	})
	r := gin.New()
	r.GET("/api/v1/admin/audit-logs/export", h.AuditLogsExport)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs/export?limit=10001", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", w.Code, w.Body.String())
	}
	if w.Header().Get("X-Export-Count") != "" {
		t.Fatalf("expected no X-Export-Count, got %q", w.Header().Get("X-Export-Count"))
	}
	if w.Header().Get("X-Export-Window") != "" {
		t.Fatalf("expected no X-Export-Window, got %q", w.Header().Get("X-Export-Window"))
	}
	if len(w.Body.String()) == 0 || w.Body.String()[0] != '{' {
		t.Fatalf("expected JSON error body, got: %q", w.Body.String())
	}
}
