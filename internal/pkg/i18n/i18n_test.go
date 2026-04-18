package i18n

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestT_NilContext(t *testing.T) {
	if T(nil, "k", "def") != "def" {
		t.Fatal()
	}
}

func TestT_NoI18nMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	if T(c, "any", "fallback") != "fallback" {
		t.Fatal()
	}
}
