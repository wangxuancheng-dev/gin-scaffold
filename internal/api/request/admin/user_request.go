package adminreq

// UserListQuery 用户列表/导出共用查询参数。
type UserListQuery struct {
	Page            int    `form:"page" json:"page"`
	PageSize        int    `form:"page_size" json:"page_size"`
	Username        string `form:"username" json:"username"`
	Nickname        string `form:"nickname" json:"nickname"`
	ExportScope     string `form:"export_scope" json:"export_scope"`           // all | page
	ExportFormat    string `form:"export_format" json:"export_format"`         // csv | xlsx
	Fields          string `form:"fields" json:"fields"`                       // id,username,nickname,created_at,role
	ExportLimit     int    `form:"export_limit" json:"export_limit"`           // only for all scope
	ExportBatchSize int    `form:"export_batch_size" json:"export_batch_size"` // stream batch size
}

// UserIDURI 后台用户路径参数。
type UserIDURI struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// UserCreateRequest 后台创建用户。
type UserCreateRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Nickname string `json:"nickname" binding:"max=64"`
	Role     string `json:"role" binding:"omitempty,oneof=admin user"`
}

// UserUpdateRequest 后台更新用户。
type UserUpdateRequest struct {
	Nickname *string `json:"nickname" binding:"omitempty,max=64"`
	Password *string `json:"password" binding:"omitempty,min=6,max=64"`
	Role     *string `json:"role" binding:"omitempty,oneof=admin user"`
}
