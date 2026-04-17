package port

import (
	"context"

	"gin-scaffold/internal/model"
)

type SystemSettingService interface {
	List(ctx context.Context, q model.SystemSettingQuery, page, pageSize int) ([]model.SystemSetting, int64, error)
	GetByID(ctx context.Context, id int64) (*model.SystemSetting, error)
	Create(ctx context.Context, key, value, valueType, groupName, remark string, actor model.SettingActor) (*model.SystemSetting, error)
	Update(ctx context.Context, id int64, value, valueType, groupName, remark *string, actor model.SettingActor) (*model.SystemSetting, error)
	Delete(ctx context.Context, id int64, actor model.SettingActor) error
	Publish(ctx context.Context, id int64, note string, actor model.SettingActor) (*model.SystemSetting, error)
	ListHistory(ctx context.Context, id int64, page, pageSize int) ([]model.SystemSettingHistory, int64, error)
	Rollback(ctx context.Context, id int64, historyID int64, reason string, actor model.SettingActor) (*model.SystemSetting, error)
}
