package adminreq

// PageQuery 通用分页请求。
type PageQuery struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}
