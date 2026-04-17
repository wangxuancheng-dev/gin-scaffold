package model

import (
	"time"

	"gorm.io/gorm"
)

type ScheduledTask struct {
	ID                int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID          string         `gorm:"size:64;not null;default:default;uniqueIndex:uk_scheduled_tasks_tenant_name,priority:1;index:idx_scheduled_tasks_tenant" json:"tenant_id"`
	Name              string         `gorm:"size:128;not null;uniqueIndex:uk_scheduled_tasks_tenant_name,priority:2" json:"name"`
	Spec              string         `gorm:"size:64;not null" json:"spec"`
	Command           string         `gorm:"size:1024;not null" json:"command"`
	TimeoutSec        int            `gorm:"not null;default:0" json:"timeout_sec"`
	ConcurrencyPolicy string         `gorm:"size:16;not null;default:forbid" json:"concurrency_policy"`
	Enabled           bool           `gorm:"not null;default:true" json:"enabled"`
	LastRunAt         *time.Time     `json:"last_run_at,omitempty"`
	LastStatus        string         `gorm:"size:32" json:"last_status,omitempty"`
	LastMessage       string         `gorm:"size:255" json:"last_message,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ScheduledTask) TableName() string {
	return "scheduled_tasks"
}

type ScheduledTaskLog struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     string         `gorm:"size:64;not null;default:default;index:idx_scheduled_task_logs_tenant_task,priority:1" json:"tenant_id"`
	TaskID       int64          `gorm:"index;not null" json:"task_id"`
	Status       string         `gorm:"size:32;not null" json:"status"`
	Output       string         `json:"output,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
	StartedAt    time.Time      `json:"started_at"`
	FinishedAt   time.Time      `json:"finished_at"`
	DurationMS   int64          `json:"duration_ms"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ScheduledTaskLog) TableName() string {
	return "scheduled_task_logs"
}
