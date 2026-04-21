package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"gin-scaffold/internal/config"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
)

type scheduledTaskRepoErrStub struct {
	getByIDErr error
	deleteErr  error
}

func (s *scheduledTaskRepoErrStub) List(context.Context, int, int) ([]model.ScheduledTask, int64, error) {
	return nil, 0, nil
}
func (s *scheduledTaskRepoErrStub) ListEnabled(context.Context) ([]model.ScheduledTask, error) {
	return nil, nil
}
func (s *scheduledTaskRepoErrStub) GetByID(context.Context, int64) (*model.ScheduledTask, error) {
	return nil, s.getByIDErr
}
func (s *scheduledTaskRepoErrStub) Create(context.Context, *model.ScheduledTask) error { return nil }
func (s *scheduledTaskRepoErrStub) Update(context.Context, *model.ScheduledTask) error { return nil }
func (s *scheduledTaskRepoErrStub) Delete(context.Context, int64) error                { return s.deleteErr }
func (s *scheduledTaskRepoErrStub) SetEnabled(context.Context, int64, bool) error      { return nil }
func (s *scheduledTaskRepoErrStub) RecordRunResult(context.Context, int64, string, string, time.Time) error {
	return nil
}
func (s *scheduledTaskRepoErrStub) AddLog(context.Context, *model.ScheduledTaskLog) error { return nil }
func (s *scheduledTaskRepoErrStub) ListLogs(context.Context, int64, int, int) ([]model.ScheduledTaskLog, int64, error) {
	return nil, 0, nil
}
func (s *scheduledTaskRepoErrStub) PurgeLogsBefore(context.Context, time.Time) error { return nil }

func TestScheduledTaskService_NotFoundMappings(t *testing.T) {
	repo := &scheduledTaskRepoErrStub{getByIDErr: gorm.ErrRecordNotFound}
	svc := NewScheduledTaskService(repo, config.SchedulerConfig{})
	ctx := context.Background()

	_, err := svc.Update(ctx, 1, nil, nil, nil, nil, nil, nil)
	assertBizNotFound(t, err)

	err = svc.SetEnabled(ctx, 1, true)
	assertBizNotFound(t, err)

	err = svc.RunNow(ctx, 1)
	assertBizNotFound(t, err)

	err = svc.ExecuteTaskByID(ctx, 1)
	assertBizNotFound(t, err)
}

func TestScheduledTaskService_DeleteNotFoundMapping(t *testing.T) {
	repo := &scheduledTaskRepoErrStub{deleteErr: gorm.ErrRecordNotFound}
	svc := NewScheduledTaskService(repo, config.SchedulerConfig{})
	err := svc.Delete(context.Background(), 1)
	assertBizNotFound(t, err)
}

func assertBizNotFound(t *testing.T, err error) {
	t.Helper()
	var biz *errcode.BizError
	if !errors.As(err, &biz) {
		t.Fatalf("expected BizError, got: %v", err)
	}
	if biz.Code != errcode.NotFound {
		t.Fatalf("expected code=%d, got=%d", errcode.NotFound, biz.Code)
	}
}
