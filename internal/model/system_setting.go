package model

import (
	"time"

	"gorm.io/gorm"
)

// SystemSetting 系统参数（键值配置）。
type SystemSetting struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  string         `gorm:"size:64;not null;default:default;uniqueIndex:uk_system_settings_tenant_key,priority:1;index:idx_system_settings_tenant" json:"tenant_id"`
	Key       string         `gorm:"size:128;not null;uniqueIndex:uk_system_settings_tenant_key,priority:2" json:"key"`
	Value     string         `gorm:"type:text;not null" json:"value"`
	ValueType string         `gorm:"size:16;not null;default:string" json:"value_type"`
	DraftValue string        `gorm:"type:text;not null" json:"draft_value"`
	DraftValueType string    `gorm:"size:16;not null;default:string" json:"draft_value_type"`
	IsPublished bool         `gorm:"not null;default:true" json:"is_published"`
	PublishedAt *time.Time   `json:"published_at"`
	PublishedBy int64        `gorm:"not null;default:0" json:"published_by"`
	PublishNote string       `gorm:"size:255;not null;default:''" json:"publish_note"`
	GroupName string         `gorm:"size:64;not null;default:''" json:"group_name"`
	Remark    string         `gorm:"size:255;not null;default:''" json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SystemSetting) TableName() string {
	return "system_settings"
}

// SystemSettingQuery 查询条件。
type SystemSettingQuery struct {
	KeyLike   string
	GroupName string
}

// SystemSettingHistory 系统参数变更历史。
type SystemSettingHistory struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SettingID      int64     `gorm:"not null;index:idx_sys_setting_hist_setting_created,priority:1" json:"setting_id"`
	SettingKey     string    `gorm:"size:128;not null;index:idx_sys_setting_hist_key_created,priority:1" json:"setting_key"`
	Action         string    `gorm:"size:16;not null" json:"action"`
	BeforeValue    string    `gorm:"type:text;not null;default:''" json:"before_value"`
	BeforeType     string    `gorm:"size:16;not null;default:''" json:"before_type"`
	BeforeGroup    string    `gorm:"size:64;not null;default:''" json:"before_group"`
	BeforeRemark   string    `gorm:"size:255;not null;default:''" json:"before_remark"`
	BeforeDeleted  bool      `gorm:"not null;default:false" json:"before_deleted"`
	AfterValue     string    `gorm:"type:text;not null;default:''" json:"after_value"`
	AfterType      string    `gorm:"size:16;not null;default:''" json:"after_type"`
	AfterGroup     string    `gorm:"size:64;not null;default:''" json:"after_group"`
	AfterRemark    string    `gorm:"size:255;not null;default:''" json:"after_remark"`
	AfterDeleted   bool      `gorm:"not null;default:false" json:"after_deleted"`
	OperatorUserID int64     `gorm:"not null;default:0" json:"operator_user_id"`
	OperatorRole   string    `gorm:"size:32;not null;default:''" json:"operator_role"`
	Reason         string    `gorm:"size:255;not null;default:''" json:"reason"`
	CreatedAt      time.Time `gorm:"index:idx_sys_setting_hist_setting_created,priority:2;index:idx_sys_setting_hist_key_created,priority:2" json:"created_at"`
}

func (SystemSettingHistory) TableName() string {
	return "system_setting_histories"
}

type SettingActor struct {
	UserID   int64
	Role     string
	TenantID string
}
