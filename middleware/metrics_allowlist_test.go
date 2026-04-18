package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMetricsAllowlist_allowsPeerInNet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, n, err := net.ParseCIDR("10.0.0.0/8")
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.Use(MetricsAllowlist("/metrics", []*net.IPNet{n}))
	r.GET("/metrics", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.RemoteAddr = "10.1.2.3:5555"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestMetricsAllowlist_blocksOutsideNet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, n, err := net.ParseCIDR("10.0.0.0/8")
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.Use(MetricsAllowlist("/metrics", []*net.IPNet{n}))
	r.GET("/metrics", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.RemoteAddr = "203.0.113.7:5555"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func TestMetricsAllowlist_skipsOtherPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, n, err := net.ParseCIDR("10.0.0.0/8")
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.Use(MetricsAllowlist("/metrics", []*net.IPNet{n}))
	r.GET("/api/v1/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/x", nil)
	req.RemoteAddr = "203.0.113.7:5555"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}
