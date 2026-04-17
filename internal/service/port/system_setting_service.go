package port

import (
	"context"

	"gin-scaffold/internal/model"
)

type SystemSettingService interface {
	List(ctx context.Context, q model.SystemSettingQuery, page, pageSize int) ([]model.SystemSetting, int64, error)
	GetByID(ctx context.Context, id int64) (*model.SystemSetting, error)
	Create(ctx context.Context, key, value, valueType, groupName, remark string) (*model.SystemSetting, error)
	Update(ctx context.Context, id int64, value, valueType, groupName, remark *string) (*model.SystemSetting, error)
	Delete(ctx context.Context, id int64) error
}
