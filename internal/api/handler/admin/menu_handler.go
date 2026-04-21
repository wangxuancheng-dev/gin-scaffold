package adminhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gin-scaffold/internal/api/handler"
	adminreq "gin-scaffold/internal/api/request/admin"
	"gin-scaffold/internal/api/response"
	adminresp "gin-scaffold/internal/api/response/admin"
	"gin-scaffold/internal/middleware"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/internal/pkg/validator"
	"gin-scaffold/internal/service/port"
)

// MenuHandler 后台菜单接口。
type MenuHandler struct {
	svc port.MenuService
}

func NewMenuHandler(s port.MenuService) *MenuHandler {
	return &MenuHandler{svc: s}
}

// ListMine 返回当前角色可见菜单（前端侧栏 / 动态路由入口）。
// @Summary 当前角色菜单
// @Tags admin-menu
// @Produce json
// @Success 200 {object} response.Body
// @Router /api/v1/admin/menus [get]
func (h *MenuHandler) ListMine(c *gin.Context) {
	claims, ok := middleware.Claims(c)
	if !ok || claims == nil {
		handler.FailUnauthorized(c, "missing claims")
		return
	}
	menus, err := h.svc.ListByRole(c.Request.Context(), claims.Role)
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	response.OK(c, gin.H{"tree": adminresp.BuildMenuTree(menus)})
}

// Catalog 返回租户下全部菜单（管理端路由表 / 菜单维护）。
// @Summary 租户菜单全量
// @Tags admin-menu
// @Produce json
// @Success 200 {object} response.Body
// @Router /api/v1/admin/menus/catalog [get]
func (h *MenuHandler) Catalog(c *gin.Context) {
	menus, err := h.svc.ListAllByTenant(c.Request.Context())
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	response.OK(c, gin.H{"tree": adminresp.BuildMenuTree(menus)})
}

// Get 菜单详情（编辑表单）。
// @Summary 菜单详情
// @Tags admin-menu
// @Produce json
// @Param id path int true "菜单ID"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/menus/{id} [get]
func (h *MenuHandler) Get(c *gin.Context) {
	var uri adminreq.MenuIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	m, err := h.svc.GetByID(c.Request.Context(), uri.ID)
	if err != nil {
		handler.FailByError(c, err, http.StatusNotFound, map[int]handler.BizMapping{
			errcode.NotFound: {Status: http.StatusNotFound},
		})
		return
	}
	response.OK(c, adminresp.FromMenu(m))
}

// Create 新建菜单（默认绑定 admin 角色可见）。
// @Summary 创建菜单
// @Tags admin-menu
// @Accept json
// @Produce json
// @Param body body adminreq.MenuCreateRequest true "参数"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/menus [post]
func (h *MenuHandler) Create(c *gin.Context) {
	var req adminreq.MenuCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	if err := validator.V().Struct(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	m, err := h.svc.Create(c.Request.Context(), req.Name, req.Path, req.PermCode, req.Sort, req.ParentID)
	if err != nil {
		handler.FailByError(c, err, http.StatusBadRequest, map[int]handler.BizMapping{
			errcode.NotFound: {Status: http.StatusNotFound},
			errcode.Conflict: {Status: http.StatusConflict},
		})
		return
	}
	response.OK(c, adminresp.FromMenu(m))
}

// Update 更新菜单。
// @Summary 更新菜单
// @Tags admin-menu
// @Accept json
// @Produce json
// @Param id path int true "菜单ID"
// @Param body body adminreq.MenuUpdateRequest true "参数"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/menus/{id} [put]
func (h *MenuHandler) Update(c *gin.Context) {
	var uri adminreq.MenuIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	var req adminreq.MenuUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	if err := validator.V().Struct(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	m, err := h.svc.Update(c.Request.Context(), uri.ID, req.Name, req.Path, req.PermCode, req.Sort, req.ParentID)
	if err != nil {
		handler.FailByError(c, err, http.StatusBadRequest, map[int]handler.BizMapping{
			errcode.NotFound:   {Status: http.StatusNotFound},
			errcode.BadRequest: {Status: http.StatusBadRequest},
			errcode.Conflict:   {Status: http.StatusConflict},
		})
		return
	}
	response.OK(c, adminresp.FromMenu(m))
}

// Delete 软删菜单。
// @Summary 删除菜单
// @Tags admin-menu
// @Produce json
// @Param id path int true "菜单ID"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/menus/{id} [delete]
func (h *MenuHandler) Delete(c *gin.Context) {
	var uri adminreq.MenuIDURI
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
