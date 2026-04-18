package clienthandler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"gin-scaffold/api/handler"
	clientreq "gin-scaffold/api/request/client"
	"gin-scaffold/api/response"
	clientresp "gin-scaffold/api/response/client"
	"gin-scaffold/internal/pkg/errcode"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/internal/pkg/validator"
	"gin-scaffold/internal/service/port"
	"gin-scaffold/middleware"
	"gin-scaffold/pkg/notify"
)

// UserHandler 客户端用户接口。
type UserHandler struct {
	svc port.UserService
}

// NewUserHandler 构造客户端用户 handler。
func NewUserHandler(s port.UserService) *UserHandler {
	return &UserHandler{svc: s}
}

// Register 用户注册。
// @Summary 用户注册
// @Tags client-user
// @Accept json
// @Produce json
// @Param body body clientreq.UserRegisterRequest true "注册参数"
// @Success 200 {object} response.Body
// @Router /api/v1/client/users [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req clientreq.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	if err := validator.V().Struct(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	u, err := h.svc.Register(c.Request.Context(), req.Username, req.Password, req.Nickname)
	if err != nil {
		handler.FailByError(c, err, http.StatusBadRequest, nil)
		return
	}
	_ = notify.Default().Notify(c.Request.Context(), notify.Message{
		Channel: "user",
		Title:   "User registered",
		Body:    u.Username,
		Meta:    map[string]string{"user_id": fmt.Sprintf("%d", u.ID)},
	})
	response.OK(c, clientresp.FromUser(u))
}

// Get 用户详情。
// @Summary 用户详情
// @Tags client-user
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Body
// @Router /api/v1/client/users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	var uri clientreq.UserIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	u, err := h.svc.GetByID(c.Request.Context(), uri.ID)
	if err != nil {
		handler.FailByError(c, err, http.StatusNotFound, map[int]handler.BizMapping{
			errcode.UserNotFound: {Status: http.StatusNotFound},
		})
		return
	}
	response.OK(c, clientresp.FromUser(u))
}

// Login 用户名密码登录，返回 JWT。
// @Summary 登录
// @Tags client-user
// @Accept json
// @Produce json
// @Param body body clientreq.LoginRequest true "登录参数"
// @Success 200 {object} response.Body
// @Router /api/v1/client/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req clientreq.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	if err := validator.V().Struct(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	access, refresh, err := h.svc.LoginWithRefresh(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		handler.FailByError(c, err, http.StatusUnauthorized, nil)
		return
	}
	response.OK(c, gin.H{"access_token": access, "refresh_token": refresh})
}

// Refresh 刷新 access token。
// @Summary 刷新令牌
// @Tags client-user
// @Accept json
// @Produce json
// @Param body body clientreq.RefreshTokenRequest true "刷新参数"
// @Success 200 {object} response.Body
// @Router /api/v1/client/auth/refresh [post]
func (h *UserHandler) Refresh(c *gin.Context) {
	var req clientreq.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	access, refresh, err := h.svc.RefreshAccess(c.Request.Context(), req.RefreshToken)
	if err != nil {
		handler.FailByError(c, err, http.StatusUnauthorized, nil)
		return
	}
	response.OK(c, gin.H{"access_token": access, "refresh_token": refresh})
}

// Logout 吊销当前 access token（加入黑名单直到过期）。
// @Summary 登出
// @Tags client-user
// @Produce json
// @Success 200 {object} response.Body
// @Router /api/v1/client/auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	claims, ok := middleware.Claims(c)
	if !ok || claims == nil || claims.ExpiresAt == nil {
		handler.FailUnauthorized(c, "missing claims")
		return
	}
	raw, ok := middleware.RawToken(c)
	if !ok || raw == "" {
		handler.FailUnauthorized(c, "missing token")
		return
	}
	if err := jwtpkg.RevokeAccessToken(c.Request.Context(), raw, claims.ExpiresAt.Time); err != nil {
		handler.FailInternal(c, err)
		return
	}
	_ = jwtpkg.ClearRefreshJTI(c.Request.Context(), claims.UserID)
	response.OK(c, gin.H{"revoked_until": claims.ExpiresAt.Time.Format(time.RFC3339)})
}
