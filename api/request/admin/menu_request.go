package adminreq

// MenuIDURI 菜单路径参数。
type MenuIDURI struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// MenuCreateRequest 创建菜单。
type MenuCreateRequest struct {
	Name     string  `json:"name" binding:"required,max=128"`
	Path     string  `json:"path" binding:"required,max=255"`
	PermCode string  `json:"perm_code" binding:"required,max=128"`
	Sort     int     `json:"sort"`
	ParentID *int64  `json:"parent_id"` // 不传或 0 表示顶级菜单
}

// MenuUpdateRequest 更新菜单（字段均为可选）。
type MenuUpdateRequest struct {
	Name     *string `json:"name" binding:"omitempty,max=128"`
	Path     *string `json:"path" binding:"omitempty,max=255"`
	PermCode *string `json:"perm_code" binding:"omitempty,max=128"`
	Sort     *int    `json:"sort"`
	ParentID *int64  `json:"parent_id"` // 传 0 表示改为顶级；不传表示不修改父级
}
