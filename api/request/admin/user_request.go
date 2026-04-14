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
