package adminresp

import (
	"slices"

	"gin-scaffold/internal/model"
)

// MenuVO 后台菜单（树形 children；前端侧栏 / 动态路由）。
type MenuVO struct {
	ID       int64                  `json:"id"`
	ParentID *int64                 `json:"parent_id,omitempty"`
	Name     string                 `json:"name"`
	Path     string                 `json:"path"`
	PermCode string                 `json:"perm_code"`
	Sort     int                    `json:"sort"`
	Meta     map[string]interface{} `json:"meta"`
	Children []*MenuVO              `json:"children,omitempty"`
}

// FromMenu 将模型转为前端菜单项。
func FromMenu(m *model.Menu) MenuVO {
	if m == nil {
		return MenuVO{}
	}
	var parentCopy *int64
	if m.ParentID != nil {
		v := *m.ParentID
		parentCopy = &v
	}
	meta := map[string]interface{}{
		"title": m.Name,
	}
	return MenuVO{
		ID:       m.ID,
		ParentID: parentCopy,
		Name:     m.Name,
		Path:     m.Path,
		PermCode: m.PermCode,
		Sort:     m.Sort,
		Meta:     meta,
	}
}

// BuildMenuTree 将扁平菜单组装为树（根节点 parent_id 为空或 0）。
func BuildMenuTree(menus []model.Menu) []*MenuVO {
	if len(menus) == 0 {
		return nil
	}
	byID := make(map[int64]*MenuVO, len(menus))
	for i := range menus {
		vo := FromMenu(&menus[i])
		ptr := new(MenuVO)
		*ptr = vo
		byID[menus[i].ID] = ptr
	}
	var roots []*MenuVO
	for i := range menus {
		id := menus[i].ID
		node := byID[id]
		pid := menus[i].ParentID
		if pid == nil || *pid <= 0 {
			roots = append(roots, node)
			continue
		}
		p := byID[*pid]
		if p == nil {
			roots = append(roots, node)
			continue
		}
		p.Children = append(p.Children, node)
	}
	sortMenuTree(roots)
	return roots
}

func sortMenuTree(nodes []*MenuVO) {
	slices.SortFunc(nodes, func(a, b *MenuVO) int {
		if a.Sort != b.Sort {
			return a.Sort - b.Sort
		}
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	for _, n := range nodes {
		if len(n.Children) > 0 {
			sortMenuTree(n.Children)
		}
	}
}
