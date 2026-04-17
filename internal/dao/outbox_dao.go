package dao

import (
	"context"
	"time"

	"gorm.io/gorm"

	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/tenant"
)

type OutboxDAO struct {
	db *gorm.DB
}

func NewOutboxDAO(db *gorm.DB) *OutboxDAO {
	return &OutboxDAO{db: db}
}

func (d *OutboxDAO) Enqueue(ctx context.Context, topic string, payload string, maxAttempts int) (*model.OutboxEvent, error) {
	row := &model.OutboxEvent{
		Topic:       topic,
		Payload:     payload,
		Status:      "pending",
		MaxAttempts: maxAttempts,
		NextRunAt:   time.Now(),
	}
	if tid := tenant.FromContext(ctx); tid != "" {
		row.TenantID = tid
	}
	if row.TenantID == "" {
		row.TenantID = "default"
	}
	if err := d.db.WithContext(ctx).Create(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

func (d *OutboxDAO) EnqueueTx(ctx context.Context, tx *gorm.DB, topic string, payload string, maxAttempts int) (*model.OutboxEvent, error) {
	if tx == nil {
		return d.Enqueue(ctx, topic, payload, maxAttempts)
	}
	row := &model.OutboxEvent{
		Topic:       topic,
		Payload:     payload,
		Status:      "pending",
		MaxAttempts: maxAttempts,
		NextRunAt:   time.Now(),
	}
	if tid := tenant.FromContext(ctx); tid != "" {
		row.TenantID = tid
	}
	if row.TenantID == "" {
		row.TenantID = "default"
	}
	if err := tx.WithContext(ctx).Create(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

func (d *OutboxDAO) FetchDue(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows []model.OutboxEvent
	err := d.db.WithContext(ctx).
		Where("status = ? AND next_run_at <= ? AND attempts < max_attempts", "pending", time.Now()).
		Order("id asc").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (d *OutboxDAO) MarkPublished(ctx context.Context, id int64) error {
	now := time.Now()
	return d.db.WithContext(ctx).Model(&model.OutboxEvent{}).Where("id = ?", id).Updates(map[string]any{
		"status":       "published",
		"published_at": &now,
		"updated_at":   now,
	}).Error
}

func (d *OutboxDAO) MarkRetry(ctx context.Context, id int64, attempts int, nextRunAt time.Time, lastError string) error {
	return d.db.WithContext(ctx).Model(&model.OutboxEvent{}).Where("id = ?", id).Updates(map[string]any{
		"attempts":    attempts,
		"next_run_at": nextRunAt,
		"last_error":  lastError,
		"updated_at":  time.Now(),
	}).Error
}

func (d *OutboxDAO) MarkDead(ctx context.Context, id int64, attempts int, lastError string) error {
	return d.db.WithContext(ctx).Model(&model.OutboxEvent{}).Where("id = ?", id).Updates(map[string]any{
		"status":     "dead",
		"attempts":   attempts,
		"last_error": lastError,
		"updated_at": time.Now(),
	}).Error
}
