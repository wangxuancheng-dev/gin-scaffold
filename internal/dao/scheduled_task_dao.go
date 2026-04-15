package dao

import (
	"context"
	"time"

	"gorm.io/gorm"

	"gin-scaffold/internal/model"
)

type ScheduledTaskDAO struct {
	db *gorm.DB
}

func NewScheduledTaskDAO(db *gorm.DB) *ScheduledTaskDAO {
	return &ScheduledTaskDAO{db: db}
}

func (d *ScheduledTaskDAO) List(ctx context.Context, page, pageSize int) ([]model.ScheduledTask, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	var total int64
	if err := d.db.WithContext(ctx).Model(&model.ScheduledTask{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.ScheduledTask
	if err := d.db.WithContext(ctx).Order("id desc").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (d *ScheduledTaskDAO) ListEnabled(ctx context.Context) ([]model.ScheduledTask, error) {
	var rows []model.ScheduledTask
	if err := d.db.WithContext(ctx).Where("enabled = ?", true).Order("id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (d *ScheduledTaskDAO) GetByID(ctx context.Context, id int64) (*model.ScheduledTask, error) {
	var t model.ScheduledTask
	if err := d.db.WithContext(ctx).First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (d *ScheduledTaskDAO) Create(ctx context.Context, t *model.ScheduledTask) error {
	return d.db.WithContext(ctx).Create(t).Error
}

func (d *ScheduledTaskDAO) Update(ctx context.Context, t *model.ScheduledTask) error {
	return d.db.WithContext(ctx).Save(t).Error
}

func (d *ScheduledTaskDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&model.ScheduledTask{}, id).Error
}

func (d *ScheduledTaskDAO) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	return d.db.WithContext(ctx).Model(&model.ScheduledTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"enabled":    enabled,
		"updated_at": time.Now(),
	}).Error
}

func (d *ScheduledTaskDAO) RecordRunResult(ctx context.Context, taskID int64, status, message string, runAt time.Time) error {
	return d.db.WithContext(ctx).Model(&model.ScheduledTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"last_run_at":  runAt,
		"last_status":  status,
		"last_message": message,
		"updated_at":   time.Now(),
	}).Error
}

func (d *ScheduledTaskDAO) AddLog(ctx context.Context, l *model.ScheduledTaskLog) error {
	return d.db.WithContext(ctx).Create(l).Error
}

func (d *ScheduledTaskDAO) ListLogs(ctx context.Context, taskID int64, page, pageSize int) ([]model.ScheduledTaskLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	tx := d.db.WithContext(ctx).Model(&model.ScheduledTaskLog{}).Where("task_id = ?", taskID)
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.ScheduledTaskLog
	if err := d.db.WithContext(ctx).Where("task_id = ?", taskID).Order("id desc").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (d *ScheduledTaskDAO) PurgeLogsBefore(ctx context.Context, before time.Time) error {
	return d.db.WithContext(ctx).Where("created_at < ?", before).Delete(&model.ScheduledTaskLog{}).Error
}
