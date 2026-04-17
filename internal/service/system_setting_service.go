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
	GetByIDAny(ctx context.Context, id int64) (*model.SystemSetting, error)
	GetByKey(ctx context.Context, key string) (*model.SystemSetting, error)
	List(ctx context.Context, q model.SystemSettingQuery, offset, limit int) ([]model.SystemSetting, int64, error)
	Update(ctx context.Context, id int64, updates map[string]any) (*model.SystemSetting, error)
	Restore(ctx context.Context, id int64, updates map[string]any) (*model.SystemSetting, error)
	SoftDelete(ctx context.Context, id int64) error
	CreateHistory(ctx context.Context, row *model.SystemSettingHistory) error
	GetHistoryByID(ctx context.Context, id int64) (*model.SystemSettingHistory, error)
	ListHistory(ctx context.Context, settingID int64, offset, limit int) ([]model.SystemSettingHistory, int64, error)
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

func (s *SystemSettingService) Create(ctx context.Context, key, value, valueType, groupName, remark string, actor model.SettingActor) (*model.SystemSetting, error) {
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
	_ = s.recordHistory(ctx, row.ID, row.Key, "create", nil, row, "", actor)
	_ = s.cacheSet(ctx, row)
	return row, nil
}

func (s *SystemSettingService) Update(ctx context.Context, id int64, value, valueType, groupName, remark *string, actor model.SettingActor) (*model.SystemSetting, error) {
	before, err := s.dao.GetByID(ctx, id)
	if err != nil {
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
	_ = s.recordHistory(ctx, row.ID, row.Key, "update", before, row, "", actor)
	_ = s.cacheSet(ctx, row)
	return row, nil
}

func (s *SystemSettingService) Delete(ctx context.Context, id int64, actor model.SettingActor) error {
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
	_ = s.recordHistory(ctx, row.ID, row.Key, "delete", row, nil, "", actor)
	_ = s.cacheDel(ctx, row.Key)
	return nil
}

func (s *SystemSettingService) ListHistory(ctx context.Context, id int64, page, pageSize int) ([]model.SystemSettingHistory, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.dao.ListHistory(ctx, id, (page-1)*pageSize, pageSize)
}

func (s *SystemSettingService) Rollback(ctx context.Context, id int64, historyID int64, reason string, actor model.SettingActor) (*model.SystemSetting, error) {
	his, err := s.dao.GetHistoryByID(ctx, historyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.NotFound, errcode.KeyNotFound)
		}
		return nil, err
	}
	if his.SettingID != id {
		return nil, errcode.New(errcode.BadRequest, errcode.KeyInvalidParam)
	}

	current, err := s.dao.GetByIDAny(ctx, id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	updates := map[string]any{
		"value":      his.BeforeValue,
		"value_type": normalizeSettingType(his.BeforeType),
		"group_name": strings.TrimSpace(his.BeforeGroup),
		"remark":     strings.TrimSpace(his.BeforeRemark),
		"updated_at": time.Now(),
	}

	if his.BeforeDeleted {
		if err := s.dao.SoftDelete(ctx, id); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		_ = s.recordHistory(ctx, id, his.SettingKey, "rollback", current, nil, strings.TrimSpace(reason), actor)
		_ = s.cacheDel(ctx, his.SettingKey)
		return nil, nil
	}

	var row *model.SystemSetting
	if current == nil || current.ID == 0 {
		return nil, errcode.New(errcode.NotFound, errcode.KeyNotFound)
	}
	if current.DeletedAt.Valid {
		row, err = s.dao.Restore(ctx, id, updates)
	} else {
		row, err = s.dao.Update(ctx, id, updates)
	}
	if err != nil {
		return nil, err
	}
	_ = s.recordHistory(ctx, row.ID, row.Key, "rollback", current, row, strings.TrimSpace(reason), actor)
	_ = s.cacheSet(ctx, row)
	return row, nil
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

func (s *SystemSettingService) recordHistory(ctx context.Context, settingID int64, settingKey, action string, before, after *model.SystemSetting, reason string, actor model.SettingActor) error {
	if s == nil || s.dao == nil {
		return nil
	}
	row := &model.SystemSettingHistory{
		SettingID:      settingID,
		SettingKey:     strings.TrimSpace(settingKey),
		Action:         strings.TrimSpace(action),
		OperatorUserID: actor.UserID,
		OperatorRole:   strings.TrimSpace(actor.Role),
		Reason:         strings.TrimSpace(reason),
	}
	if before != nil {
		row.BeforeValue = before.Value
		row.BeforeType = before.ValueType
		row.BeforeGroup = before.GroupName
		row.BeforeRemark = before.Remark
		row.BeforeDeleted = before.DeletedAt.Valid
	}
	if after != nil {
		row.AfterValue = after.Value
		row.AfterType = after.ValueType
		row.AfterGroup = after.GroupName
		row.AfterRemark = after.Remark
		row.AfterDeleted = after.DeletedAt.Valid
	} else {
		row.AfterDeleted = true
	}
	if row.SettingKey == "" {
		if before != nil {
			row.SettingKey = before.Key
		} else if after != nil {
			row.SettingKey = after.Key
		}
	}
	return s.dao.CreateHistory(ctx, row)
}
