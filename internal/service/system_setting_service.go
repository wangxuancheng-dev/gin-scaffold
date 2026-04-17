package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/pkg/cache"
)

type SystemSettingRepo interface {
	Create(ctx context.Context, row *model.SystemSetting) error
	GetByID(ctx context.Context, id int64) (*model.SystemSetting, error)
	GetByKey(ctx context.Context, key string) (*model.SystemSetting, error)
	List(ctx context.Context, q model.SystemSettingQuery, offset, limit int) ([]model.SystemSetting, int64, error)
	Update(ctx context.Context, id int64, updates map[string]any) (*model.SystemSetting, error)
	SoftDelete(ctx context.Context, id int64) error
}

type SystemSettingService struct {
	dao   SystemSettingRepo
	cache *cache.Client
}

func NewSystemSettingService(d SystemSettingRepo) *SystemSettingService {
	return &SystemSettingService{dao: d, cache: cache.NewFromConfig()}
}

func (s *SystemSettingService) List(ctx context.Context, q model.SystemSettingQuery, page, pageSize int) ([]model.SystemSetting, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.dao.List(ctx, q, (page-1)*pageSize, pageSize)
}

func (s *SystemSettingService) GetByID(ctx context.Context, id int64) (*model.SystemSetting, error) {
	row, err := s.dao.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.NotFound, errcode.KeyNotFound)
		}
		return nil, err
	}
	return row, nil
}

func (s *SystemSettingService) Create(ctx context.Context, key, value, valueType, groupName, remark string) (*model.SystemSetting, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, errcode.New(errcode.BadRequest, errcode.KeyInvalidParam)
	}
	row := &model.SystemSetting{
		Key:       key,
		Value:     value,
		ValueType: normalizeSettingType(valueType),
		GroupName: strings.TrimSpace(groupName),
		Remark:    strings.TrimSpace(remark),
	}
	if err := s.dao.Create(ctx, row); err != nil {
		return nil, err
	}
	_ = s.cacheSet(ctx, row)
	return row, nil
}

func (s *SystemSettingService) Update(ctx context.Context, id int64, value, valueType, groupName, remark *string) (*model.SystemSetting, error) {
	if _, err := s.dao.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.NotFound, errcode.KeyNotFound)
		}
		return nil, err
	}
	updates := map[string]any{
		"updated_at": time.Now(),
	}
	if value != nil {
		updates["value"] = *value
	}
	if valueType != nil {
		updates["value_type"] = normalizeSettingType(*valueType)
	}
	if groupName != nil {
		updates["group_name"] = strings.TrimSpace(*groupName)
	}
	if remark != nil {
		updates["remark"] = strings.TrimSpace(*remark)
	}
	row, err := s.dao.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}
	_ = s.cacheSet(ctx, row)
	return row, nil
}

func (s *SystemSettingService) Delete(ctx context.Context, id int64) error {
	row, err := s.dao.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.New(errcode.NotFound, errcode.KeyNotFound)
		}
		return err
	}
	if err := s.dao.SoftDelete(ctx, id); err != nil {
		return err
	}
	_ = s.cacheDel(ctx, row.Key)
	return nil
}

func normalizeSettingType(in string) string {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case "int", "bool", "json":
		return strings.ToLower(strings.TrimSpace(in))
	default:
		return "string"
	}
}

func (s *SystemSettingService) cacheKey(key string) string {
	if s == nil || s.cache == nil {
		return ""
	}
	return s.cache.Key("sys_setting", strings.TrimSpace(key))
}

func (s *SystemSettingService) cacheSet(ctx context.Context, row *model.SystemSetting) error {
	if s == nil || s.cache == nil || row == nil || strings.TrimSpace(row.Key) == "" {
		return nil
	}
	return s.cache.SetJSON(ctx, s.cacheKey(row.Key), row, 10*time.Minute)
}

func (s *SystemSettingService) cacheDel(ctx context.Context, key string) error {
	if s == nil || s.cache == nil || strings.TrimSpace(key) == "" {
		return nil
	}
	return s.cache.Del(ctx, s.cacheKey(key))
}
