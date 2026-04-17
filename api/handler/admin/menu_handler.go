package adminhandler

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"

	"gin-scaffold/api/handler"
	"gin-scaffold/api/response"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/service/port"
	"gin-scaffold/middleware"
)

// MenuHandler 后台菜单接口。
type MenuHandler struct {
	svc port.MenuService
}

func NewMenuHandler(s port.MenuService) *MenuHandler {
	return &MenuHandler{svc: s}
}

// ListMine 返回当前角色可见菜单。
// @Summary 菜单列表（后台）
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
	data := lo.Map(menus, func(m model.Menu, _ int) gin.H {
		return gin.H{
			"id":        m.ID,
			"name":      m.Name,
			"path":      m.Path,
			"perm_code": m.PermCode,
			"sort":      m.Sort,
		}
	})
	response.OK(c, gin.H{"list": data})
}
