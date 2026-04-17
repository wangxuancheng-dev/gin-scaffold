package adminreq

// AuditLogListQuery 审计日志分页查询。
type AuditLogListQuery struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	UserID    int64  `form:"user_id"`
	Action    string `form:"action"` // POST|PUT|PATCH|DELETE
	Status    int    `form:"status"`
	Path      string `form:"path"`
	RequestID string `form:"request_id"`
	From      string `form:"from"` // RFC3339
	To        string `form:"to"`   // RFC3339
}
