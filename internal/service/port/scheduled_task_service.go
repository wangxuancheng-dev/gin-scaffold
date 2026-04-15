package port

import (
	"context"

	"gin-scaffold/internal/model"
)

type ScheduledTaskService interface {
	List(ctx context.Context, page, pageSize int) ([]model.ScheduledTask, int64, error)
	Create(ctx context.Context, name, spec, command string, timeoutSec int, concurrencyPolicy string, enabled bool) (*model.ScheduledTask, error)
	Update(ctx context.Context, id int64, name, spec, command *string, timeoutSec *int, concurrencyPolicy *string, enabled *bool) (*model.ScheduledTask, error)
	Delete(ctx context.Context, id int64) error
	SetEnabled(ctx context.Context, id int64, enabled bool) error
	RunNow(ctx context.Context, id int64) error
	ListLogs(ctx context.Context, taskID int64, page, pageSize int) ([]model.ScheduledTaskLog, int64, error)
}
