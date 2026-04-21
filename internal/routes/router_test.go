package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"gin-scaffold/internal/api/handler"
	"gin-scaffold/internal/config"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/pkg/limiter"
)

func TestMain(m *testing.M) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
	if err := os.Chdir(root); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestTruncatePath(t *testing.T) {
	got := truncatePath("hi", 6)
	if len(got) != 6 {
		t.Fatalf("len=%d %q", len(got), got)
	}
	if !strings.HasPrefix(truncatePath("abcdefghijklmnop", 8), "abcde") {
		t.Fatal()
	}
}

func TestIsLoopbackClient(t *testing.T) {
	if !isLoopbackClient("127.0.0.1") {
		t.Fatal()
	}
	if isLoopbackClient("192.0.2.1") {
		t.Fatal()
	}
	if isLoopbackClient("not-an-ip") {
		t.Fatal()
	}
}

func TestBuild_LivezRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.App{
		Env:   "test",
		Name:  "router-test",
		Debug: false,
		HTTP: config.HTTPConfig{
			Host:              "0.0.0.0",
			Port:              8080,
			ReadTimeout:       5,
			ReadHeaderTimeout: 5,
			WriteTimeout:      5,
			IdleTimeout:       5,
			ShutdownTimeout:   5,
			MaxBodyBytes:      1 << 20,
			SwaggerEnabled:    false,
		},
		Metrics: config.MetricsConfig{Enabled: false},
		JWT: config.JWTConfig{
			Secret:           "unit-test-router-secret",
			AccessExpireMin:  60,
			RefreshExpireMin: 1440,
			Issuer:           "router-test",
		},
		I18n: config.I18nConfig{
			DefaultLang: "zh",
			BundlePaths: []string{"./i18n/zh.json"},
		},
		Limiter: config.LimiterConfig{
			IPRPS: 1000, IPBurst: 1000, RouteRPS: 2000, RouteBurst: 2000,
		},
		CORS:   config.CORSConfig{AllowOrigins: []string{"*"}},
		Tenant: config.TenantConfig{Enabled: false},
	}
	jm := jwtpkg.NewManager(&cfg.JWT)
	lim := limiter.NewStore(cfg.Limiter.IPRPS, cfg.Limiter.IPBurst, cfg.Limiter.RouteRPS, cfg.Limiter.RouteBurst)
	e, err := Build(Options{
		Cfg:     cfg,
		JWT:     jm,
		Base:    &handler.BaseHandler{},
		Limiter: lim,
	})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	e.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestBuild_metricsAllowlistInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.App{
		Env:   "test",
		Name:  "router-test",
		Debug: false,
		HTTP: config.HTTPConfig{
			Host:              "0.0.0.0",
			Port:              8080,
			ReadTimeout:       5,
			ReadHeaderTimeout: 5,
			WriteTimeout:      5,
			IdleTimeout:       5,
			ShutdownTimeout:   5,
			MaxBodyBytes:      1 << 20,
			SwaggerEnabled:    false,
		},
		Metrics: config.MetricsConfig{
			Enabled:         true,
			Path:            "/metrics",
			AllowedNetworks: []string{"not-a-cidr"},
		},
		JWT: config.JWTConfig{
			Secret:           "unit-test-router-secret",
			AccessExpireMin:  60,
			RefreshExpireMin: 1440,
			Issuer:           "router-test",
		},
		I18n: config.I18nConfig{
			DefaultLang: "zh",
			BundlePaths: []string{"./i18n/zh.json"},
		},
		Limiter: config.LimiterConfig{
			IPRPS: 1000, IPBurst: 1000, RouteRPS: 2000, RouteBurst: 2000,
		},
		CORS:   config.CORSConfig{AllowOrigins: []string{"*"}},
		Tenant: config.TenantConfig{Enabled: false},
	}
	jm := jwtpkg.NewManager(&cfg.JWT)
	lim := limiter.NewStore(cfg.Limiter.IPRPS, cfg.Limiter.IPBurst, cfg.Limiter.RouteRPS, cfg.Limiter.RouteBurst)
	_, err := Build(Options{
		Cfg:     cfg,
		JWT:     jm,
		Base:    &handler.BaseHandler{},
		Limiter: lim,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "metrics allowlist")
}
