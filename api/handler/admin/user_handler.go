package adminhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"

	"gin-scaffold/api/handler/userapi"
	adminreq "gin-scaffold/api/request/admin"
	"gin-scaffold/api/response"
	clientresp "gin-scaffold/api/response/client"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
)

// UserHandler 后台用户接口。
type UserHandler struct {
	svc userapi.Service
}

// NewUserHandler 构造后台用户 handler。
func NewUserHandler(s userapi.Service) *UserHandler {
	return &UserHandler{svc: s}
}

// List 用户分页（后台）。
// @Summary 用户列表（后台）
// @Tags admin-user
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/users [get]
func (h *UserHandler) List(c *gin.Context) {
	var q adminreq.PageQuery
	_ = c.ShouldBindQuery(&q)
	rows, total, err := h.svc.List(c.Request.Context(), q.Page, q.PageSize)
	if err != nil {
		response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, err.Error())
		return
	}
	list := lo.Map(rows, func(u model.User, _ int) clientresp.UserVO {
		return clientresp.FromUser(&u)
	})
	response.OK(c, gin.H{"total": total, "list": list})
}
