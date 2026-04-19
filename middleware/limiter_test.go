package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"gin-scaffold/pkg/limiter"
)

func TestLimiterWithBackendKeys_defaultMatchesLimiterWithBackend(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lim := limiter.NewStore(100, 100, 100, 100)
	e := gin.New()
	e.Use(LimiterWithBackendKeys(lim, nil))
	e.GET("/p", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/p", nil))
	require.Equal(t, http.StatusOK, w.Code)
}

func TestLimiterWithBackendKeys_customRouteKey_separateBuckets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// 路由维极严：默认同一 route key 下第二次即 429
	lim := limiter.NewStore(1000, 1000, 0.0001, 1)
	e := gin.New()
	e.Use(LimiterWithBackendKeys(lim, &LimiterKeys{
		RouteKey: func(c *gin.Context) string {
			return "bucket:" + c.GetHeader("X-Bucket")
		},
	}))
	e.GET("/api", func(c *gin.Context) { c.Status(http.StatusOK) })

	reqA := httptest.NewRequest(http.MethodGet, "/api", nil)
	reqA.Header.Set("X-Bucket", "a")
	w1 := httptest.NewRecorder()
	e.ServeHTTP(w1, reqA)
	require.Equal(t, http.StatusOK, w1.Code)

	reqB := httptest.NewRequest(http.MethodGet, "/api", nil)
	reqB.Header.Set("X-Bucket", "b")
	w2 := httptest.NewRecorder()
	e.ServeHTTP(w2, reqB)
	require.Equal(t, http.StatusOK, w2.Code, "different custom route keys should not share the same route bucket")
}

func TestLimiterWithBackendKeys_customIPKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lim := limiter.NewStore(0.0001, 1, 1000, 1000)
	e := gin.New()
	e.Use(LimiterWithBackendKeys(lim, &LimiterKeys{
		IPKey: func(c *gin.Context) string {
			return "ip-scope:" + c.GetHeader("X-Scope")
		},
	}))
	e.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	r1 := httptest.NewRequest(http.MethodGet, "/x", nil)
	r1.Header.Set("X-Real-IP", "203.0.113.1")
	r1.Header.Set("X-Scope", "1")
	w1 := httptest.NewRecorder()
	e.ServeHTTP(w1, r1)
	require.Equal(t, http.StatusOK, w1.Code)

	r2 := httptest.NewRequest(http.MethodGet, "/x", nil)
	r2.Header.Set("X-Real-IP", "203.0.113.1")
	r2.Header.Set("X-Scope", "2")
	w2 := httptest.NewRecorder()
	e.ServeHTTP(w2, r2)
	require.Equal(t, http.StatusOK, w2.Code, "same TCP client IP but different custom IP key should use separate IP buckets")
}
