package model

import "time"

type OutboxEvent struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    string     `gorm:"size:64;not null;default:default;index:idx_outbox_tenant_status_next,priority:1" json:"tenant_id"`
	Topic       string     `gorm:"size:128;not null;index:idx_outbox_topic_created,priority:1" json:"topic"`
	Payload     string     `gorm:"type:longtext;not null" json:"payload"`
	Status      string     `gorm:"size:16;not null;default:pending;index:idx_outbox_tenant_status_next,priority:2" json:"status"`
	Attempts    int        `gorm:"not null;default:0" json:"attempts"`
	MaxAttempts int        `gorm:"not null;default:10" json:"max_attempts"`
	NextRunAt   time.Time  `gorm:"not null;index:idx_outbox_tenant_status_next,priority:3" json:"next_run_at"`
	LastError   string     `gorm:"size:512;not null;default:''" json:"last_error"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `gorm:"index:idx_outbox_topic_created,priority:2" json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (OutboxEvent) TableName() string {
	return "outbox_events"
}
