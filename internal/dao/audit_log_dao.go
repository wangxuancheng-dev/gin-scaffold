package dao

import (
	"context"

	"gorm.io/gorm"

	"gin-scaffold/internal/model"
)

// AuditLogDAO 审计日志写入。
type AuditLogDAO struct {
	db *gorm.DB
}

// NewAuditLogDAO 构造。
func NewAuditLogDAO(db *gorm.DB) *AuditLogDAO {
	return &AuditLogDAO{db: db}
}

// Create 插入一条审计记录。
func (d *AuditLogDAO) Create(ctx context.Context, row *model.AuditLog) error {
	if d == nil || d.db == nil || row == nil {
		return nil
	}
	return d.db.WithContext(ctx).Create(row).Error
}
