package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"

	"gin-scaffold/api/request"
	"gin-scaffold/api/response"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/internal/pkg/validator"
	"gin-scaffold/middleware"
)

// UserServiceAPI 定义 UserHandler 依赖的服务接口，便于测试注入 mock。
type UserServiceAPI interface {
	Register(ctx context.Context, username, password, nickname string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	Login(ctx context.Context, username, password string) (string, error)
	LoginWithRefresh(ctx context.Context, username, password string) (string, string, error)
	RefreshAccess(ctx context.Context, refreshToken string) (string, string, error)
	List(ctx context.Context, page, pageSize int) ([]model.User, int64, error)
}

// UserHandler 用户 HTTP 层。
type UserHandler struct {
	svc UserServiceAPI
}

// NewUserHandler 构造。
func NewUserHandler(s UserServiceAPI) *UserHandler {
	return &UserHandler{svc: s}
}

// Register 用户注册。
// @Summary 用户注册
// @Tags user
// @Accept json
// @Produce json
// @Param body body request.UserRegisterRequest true "注册参数"
// @Success 200 {object} response.Body
// @Router /api/v1/client/users [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req request.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	if err := validator.V().Struct(&req); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	u, err := h.svc.Register(c.Request.Context(), req.Username, req.Password, req.Nickname)
	if err != nil {
		var biz *errcode.BizError
		if errors.As(err, &biz) {
			response.FailBiz(c, biz.Code, biz.Key, biz.Error())
			return
		}
		response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, err.Error())
		return
	}
	response.OK(c, response.FromUser(u))
}

// Get 用户详情。
// @Summary 用户详情
// @Tags user
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Body
// @Router /api/v1/client/users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	var uri request.UserIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	u, err := h.svc.GetByID(c.Request.Context(), uri.ID)
	if err != nil {
		var biz *errcode.BizError
		if errors.As(err, &biz) {
			response.FailHTTP(c, http.StatusNotFound, biz.Code, biz.Key, biz.Error())
			return
		}
		response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, err.Error())
		return
	}
	response.OK(c, response.FromUser(u))
}

// Login 用户名密码登录，返回 JWT。
// @Summary 登录
// @Tags user
// @Accept json
// @Produce json
// @Param body body request.LoginRequest true "登录参数"
// @Success 200 {object} response.Body
// @Router /api/v1/client/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	if err := validator.V().Struct(&req); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	access, refresh, err := h.svc.LoginWithRefresh(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		var biz *errcode.BizError
		if errors.As(err, &biz) {
			response.FailHTTP(c, http.StatusUnauthorized, biz.Code, biz.Key, biz.Error())
			return
		}
		response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, err.Error())
		return
	}
	response.OK(c, gin.H{"access_token": access, "refresh_token": refresh})
}

// Refresh 刷新 access token。
// @Summary 刷新令牌
// @Tags user
// @Accept json
// @Produce json
// @Param body body request.RefreshTokenRequest true "刷新参数"
// @Success 200 {object} response.Body
// @Router /api/v1/client/auth/refresh [post]
func (h *UserHandler) Refresh(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	access, refresh, err := h.svc.RefreshAccess(c.Request.Context(), req.RefreshToken)
	if err != nil {
		var biz *errcode.BizError
		if errors.As(err, &biz) {
			response.FailHTTP(c, http.StatusUnauthorized, biz.Code, biz.Key, biz.Error())
			return
		}
		response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, err.Error())
		return
	}
	response.OK(c, gin.H{"access_token": access, "refresh_token": refresh})
}

// Logout 吊销当前 access token（加入黑名单直到过期）。
// @Summary 登出
// @Tags user
// @Produce json
// @Success 200 {object} response.Body
// @Router /api/v1/client/auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	claims, ok := middleware.Claims(c)
	if !ok || claims == nil || claims.ExpiresAt == nil {
		response.FailHTTP(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.KeyUnauthorized, "missing claims")
		return
	}
	raw, ok := middleware.RawToken(c)
	if !ok || raw == "" {
		response.FailHTTP(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.KeyUnauthorized, "missing token")
		return
	}
	if err := jwtpkg.RevokeAccessToken(c.Request.Context(), raw, claims.ExpiresAt.Time); err != nil {
		response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, err.Error())
		return
	}
	response.OK(c, gin.H{"revoked_until": claims.ExpiresAt.Time.Format(time.RFC3339)})
}

// List 用户分页。
// @Summary 用户列表
// @Tags user
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} response.Body
// @Router /api/v1/client/users [get]
func (h *UserHandler) List(c *gin.Context) {
	var q request.PageQuery
	_ = c.ShouldBindQuery(&q)
	rows, total, err := h.svc.List(c.Request.Context(), q.Page, q.PageSize)
	if err != nil {
		response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, err.Error())
		return
	}
	list := lo.Map(rows, func(u model.User, _ int) response.UserVO {
		return response.FromUser(&u)
	})
	response.OK(c, gin.H{"total": total, "list": list})
}
