package model

import "time"

// AuditLog HTTP 写操作审计（不含请求体，避免敏感数据落库）。
type AuditLog struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestID   string    `gorm:"size:64;index:idx_audit_request" json:"request_id"`
	UserID      int64     `gorm:"index:idx_audit_user_created,priority:1" json:"user_id"`
	Role        string    `gorm:"size:32" json:"role"`
	ActorType   string    `gorm:"size:16" json:"actor_type"`
	Action      string    `gorm:"size:16;not null" json:"action"`
	Path        string    `gorm:"size:512;not null" json:"path"`
	Query       string    `gorm:"size:1024" json:"query"`
	Status      int       `gorm:"not null" json:"status"`
	LatencyMS   int       `gorm:"not null" json:"latency_ms"`
	ClientIP    string    `gorm:"size:64" json:"client_ip"`
	CreatedAt   time.Time `gorm:"index:idx_audit_user_created,priority:2;index:idx_audit_created" json:"created_at"`
}

// TableName 表名。
func (AuditLog) TableName() string {
	return "audit_logs"
}
