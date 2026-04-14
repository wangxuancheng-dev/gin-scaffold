package unit_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"gin-scaffold/api/handler"
	"gin-scaffold/internal/pkg/errcode"
)

type errResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func TestUserHandler_Register_BadRequest(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	h := handler.NewUserHandler(nil)
	r := gin.New()
	r.POST("/users", h.Register)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"username":"ab"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var resp errResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, errcode.BadRequest, resp.Code)
}

func TestUserHandler_Login_BadRequest(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	h := handler.NewUserHandler(nil)
	r := gin.New()
	r.POST("/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"username":"u","password":"1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var resp errResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, errcode.BadRequest, resp.Code)
}
