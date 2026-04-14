package model

import "time"

// UserExportRow 用户导出行。
type UserExportRow struct {
	ID        int64
	Username  string
	Nickname  string
	CreatedAt time.Time
	Role      string
}
