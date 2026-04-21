package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBaseHandler_Livez(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &BaseHandler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/livez", nil)
	h.Livez(c)
	if w.Code != 200 {
		t.Fatalf("code=%d", w.Code)
	}
}
