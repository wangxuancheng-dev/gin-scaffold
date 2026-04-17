package model

import (
	"time"

	"gorm.io/gorm"
)

// SystemSetting 系统参数（键值配置）。
type SystemSetting struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Key       string         `gorm:"size:128;uniqueIndex;not null" json:"key"`
	Value     string         `gorm:"type:text;not null" json:"value"`
	ValueType string         `gorm:"size:16;not null;default:string" json:"value_type"`
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
