package adminhandler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"gin-scaffold/api/handler"
	adminreq "gin-scaffold/api/request/admin"
	"gin-scaffold/api/response"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/internal/pkg/validator"
	"gin-scaffold/internal/service/port"
)

type SystemSettingHandler struct {
	svc port.SystemSettingService
}

func NewSystemSettingHandler(s port.SystemSettingService) *SystemSettingHandler {
	return &SystemSettingHandler{svc: s}
}

// List 系统参数分页（后台）。
// @Summary 系统参数列表（后台）
// @Tags admin-system-setting
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Param key query string false "键名模糊匹配"
// @Param group_name query string false "分组名精确匹配"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/system-settings [get]
func (h *SystemSettingHandler) List(c *gin.Context) {
	var q adminreq.SystemSettingListQuery
	_ = c.ShouldBindQuery(&q)
	rows, total, err := h.svc.List(c.Request.Context(), model.SystemSettingQuery{
		KeyLike:   strings.TrimSpace(q.Key),
		GroupName: strings.TrimSpace(q.GroupName),
	}, q.Page, q.PageSize)
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	response.OK(c, gin.H{"total": total, "list": rows})
}

// Get 系统参数详情（后台）。
// @Summary 系统参数详情（后台）
// @Tags admin-system-setting
// @Produce json
// @Param id path int true "参数ID"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/system-settings/{id} [get]
func (h *SystemSettingHandler) Get(c *gin.Context) {
	var uri adminreq.SystemSettingIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	row, err := h.svc.GetByID(c.Request.Context(), uri.ID)
	if err != nil {
		handler.FailByError(c, err, http.StatusNotFound, map[int]handler.BizMapping{
			errcode.NotFound: {Status: http.StatusNotFound},
		})
		return
	}
	response.OK(c, row)
}

// Create 新增系统参数（后台）。
// @Summary 新增系统参数（后台）
// @Tags admin-system-setting
// @Accept json
// @Produce json
// @Param body body adminreq.SystemSettingCreateRequest true "创建参数"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/system-settings [post]
func (h *SystemSettingHandler) Create(c *gin.Context) {
	var req adminreq.SystemSettingCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	if err := validator.V().Struct(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	row, err := h.svc.Create(c.Request.Context(), req.Key, req.Value, req.ValueType, req.GroupName, req.Remark)
	if err != nil {
		handler.FailByError(c, err, http.StatusBadRequest, nil)
		return
	}
	response.OK(c, row)
}

// Update 更新系统参数（后台）。
// @Summary 更新系统参数（后台）
// @Tags admin-system-setting
// @Accept json
// @Produce json
// @Param id path int true "参数ID"
// @Param body body adminreq.SystemSettingUpdateRequest true "更新参数"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/system-settings/{id} [put]
func (h *SystemSettingHandler) Update(c *gin.Context) {
	var uri adminreq.SystemSettingIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	var req adminreq.SystemSettingUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	if err := validator.V().Struct(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	row, err := h.svc.Update(c.Request.Context(), uri.ID, req.Value, req.ValueType, req.GroupName, req.Remark)
	if err != nil {
		handler.FailByError(c, err, http.StatusBadRequest, map[int]handler.BizMapping{
			errcode.NotFound: {Status: http.StatusNotFound},
		})
		return
	}
	response.OK(c, row)
}

// Delete 删除系统参数（后台）。
// @Summary 删除系统参数（后台）
// @Tags admin-system-setting
// @Produce json
// @Param id path int true "参数ID"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/system-settings/{id} [delete]
func (h *SystemSettingHandler) Delete(c *gin.Context) {
	var uri adminreq.SystemSettingIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uri.ID); err != nil {
		handler.FailByError(c, err, http.StatusBadRequest, map[int]handler.BizMapping{
			errcode.NotFound: {Status: http.StatusNotFound},
		})
		return
	}
	response.OK(c, gin.H{"deleted": true})
}
