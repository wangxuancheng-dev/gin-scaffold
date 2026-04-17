package dao

import (
	"context"

	"gorm.io/gorm"

	"gin-scaffold/internal/model"
)

type SystemSettingDAO struct {
	db *gorm.DB
}

func NewSystemSettingDAO(db *gorm.DB) *SystemSettingDAO {
	return &SystemSettingDAO{db: db}
}

func (d *SystemSettingDAO) Create(ctx context.Context, row *model.SystemSetting) error {
	return d.db.WithContext(ctx).Create(row).Error
}

func (d *SystemSettingDAO) GetByID(ctx context.Context, id int64) (*model.SystemSetting, error) {
	var row model.SystemSetting
	if err := d.db.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (d *SystemSettingDAO) GetByKey(ctx context.Context, key string) (*model.SystemSetting, error) {
	var row model.SystemSetting
	if err := d.db.WithContext(ctx).Where("`key` = ?", key).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (d *SystemSettingDAO) applyFilters(tx *gorm.DB, q model.SystemSettingQuery) *gorm.DB {
	if q.KeyLike != "" {
		tx = tx.Where("`key` LIKE ?", "%"+q.KeyLike+"%")
	}
	if q.GroupName != "" {
		tx = tx.Where("group_name = ?", q.GroupName)
	}
	return tx
}

func (d *SystemSettingDAO) List(ctx context.Context, q model.SystemSettingQuery, offset, limit int) ([]model.SystemSetting, int64, error) {
	tx := d.applyFilters(d.db.WithContext(ctx).Model(&model.SystemSetting{}), q)
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.SystemSetting
	if err := d.applyFilters(d.db.WithContext(ctx), q).Order("id desc").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (d *SystemSettingDAO) Update(ctx context.Context, id int64, updates map[string]any) (*model.SystemSetting, error) {
	if err := d.db.WithContext(ctx).Model(&model.SystemSetting{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return d.GetByID(ctx, id)
}

func (d *SystemSettingDAO) SoftDelete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&model.SystemSetting{}, id).Error
}
