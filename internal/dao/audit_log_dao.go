package dao

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"gin-scaffold/internal/model"
)

// AuditLogListQuery 审计日志查询条件。
type AuditLogListQuery struct {
	Page      int
	PageSize  int
	UserID    int64
	Action    string
	Status    int
	PathLike  string
	RequestID string
	From      *time.Time
	To        *time.Time
}

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

// List 分页查询审计日志。
func (d *AuditLogDAO) List(ctx context.Context, q AuditLogListQuery) ([]model.AuditLog, int64, error) {
	if d == nil || d.db == nil {
		return []model.AuditLog{}, 0, nil
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 200 {
		q.PageSize = 20
	}
	query := d.db.WithContext(ctx).Model(&model.AuditLog{})
	if q.UserID > 0 {
		query = query.Where("user_id = ?", q.UserID)
	}
	if s := strings.TrimSpace(q.Action); s != "" {
		query = query.Where("action = ?", strings.ToUpper(s))
	}
	if q.Status > 0 {
		query = query.Where("status = ?", q.Status)
	}
	if s := strings.TrimSpace(q.PathLike); s != "" {
		query = query.Where("path LIKE ?", "%"+s+"%")
	}
	if s := strings.TrimSpace(q.RequestID); s != "" {
		query = query.Where("request_id = ?", s)
	}
	if q.From != nil {
		query = query.Where("created_at >= ?", *q.From)
	}
	if q.To != nil {
		query = query.Where("created_at <= ?", *q.To)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.AuditLog
	offset := (q.Page - 1) * q.PageSize
	if err := query.Order("id desc").Offset(offset).Limit(q.PageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// ListForExport 按筛选导出（最多 maxRows 条，按 id desc）。
func (d *AuditLogDAO) ListForExport(ctx context.Context, q AuditLogListQuery, maxRows int) ([]model.AuditLog, error) {
	if d == nil || d.db == nil {
		return []model.AuditLog{}, nil
	}
	if maxRows <= 0 || maxRows > 10000 {
		maxRows = 5000
	}
	query := d.db.WithContext(ctx).Model(&model.AuditLog{})
	if q.UserID > 0 {
		query = query.Where("user_id = ?", q.UserID)
	}
	if s := strings.TrimSpace(q.Action); s != "" {
		query = query.Where("action = ?", strings.ToUpper(s))
	}
	if q.Status > 0 {
		query = query.Where("status = ?", q.Status)
	}
	if s := strings.TrimSpace(q.PathLike); s != "" {
		query = query.Where("path LIKE ?", "%"+s+"%")
	}
	if s := strings.TrimSpace(q.RequestID); s != "" {
		query = query.Where("request_id = ?", s)
	}
	if q.From != nil {
		query = query.Where("created_at >= ?", *q.From)
	}
	if q.To != nil {
		query = query.Where("created_at <= ?", *q.To)
	}
	var rows []model.AuditLog
	if err := query.Order("id desc").Limit(maxRows).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
