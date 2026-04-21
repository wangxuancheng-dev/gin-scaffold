package adminhandler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"gin-scaffold/internal/dao"
	"gin-scaffold/internal/job"
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

func TestBuildExportFilterSummary(t *testing.T) {
	t.Parallel()
	from := time.Now().UTC().Add(-24 * time.Hour)
	to := time.Now().UTC()
	s := buildExportFilterSummary(dao.AuditLogListQuery{
		UserID:    12,
		Action:    "post",
		Status:    200,
		PathLike:  "/api/v1/admin",
		RequestID: "rid-1",
		From:      &from,
		To:        &to,
	})
	wantContains := []string{
		"user_id=12", "action=POST", "status=200", "path~=/api/v1/admin", "request_id=rid-1", "from=", "to=",
	}
	for _, w := range wantContains {
		if !strings.Contains(s, w) {
			t.Fatalf("filter summary should contain %q, got %q", w, s)
		}
	}
}

func TestBuildAbsoluteURL(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
	c.Request.Host = "api.example.com"
	got := buildAbsoluteURL(c, "/api/v1/admin/audit-logs/export/tasks/t1/download")
	if got != "http://api.example.com/api/v1/admin/audit-logs/export/tasks/t1/download" {
		t.Fatalf("unexpected absolute url: %s", got)
	}
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	got = buildAbsoluteURL(c, "/p")
	if got != "https://api.example.com/p" {
		t.Fatalf("unexpected forwarded proto url: %s", got)
	}
}

func TestBuildAuditExportStatusResponse_IsReady(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
	c.Request.Host = "example.com"
	notReady := buildAuditExportStatusResponse(c, &job.AuditExportStatus{
		TaskID: "t1",
		State:  "running",
	}, "t1")
	if notReady["is_ready"] != false {
		t.Fatalf("expected is_ready=false, got %v", notReady["is_ready"])
	}
	ready := buildAuditExportStatusResponse(c, &job.AuditExportStatus{
		TaskID:  "t2",
		State:   "success",
		FileKey: "exports/audit/a.csv",
	}, "t2")
	if ready["is_ready"] != true {
		t.Fatalf("expected is_ready=true, got %v", ready["is_ready"])
	}
	if !strings.Contains(ready["download_url"].(string), "/api/v1/admin/audit-logs/export/tasks/t2/download") {
		t.Fatalf("unexpected download_url: %v", ready["download_url"])
	}
}
