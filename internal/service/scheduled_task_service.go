package service

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"gorm.io/gorm"

	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
)

type ScheduledTaskRepo interface {
	List(ctx context.Context, page, pageSize int) ([]model.ScheduledTask, int64, error)
	ListEnabled(ctx context.Context) ([]model.ScheduledTask, error)
	GetByID(ctx context.Context, id int64) (*model.ScheduledTask, error)
	Create(ctx context.Context, t *model.ScheduledTask) error
	Update(ctx context.Context, t *model.ScheduledTask) error
	Delete(ctx context.Context, id int64) error
	SetEnabled(ctx context.Context, id int64, enabled bool) error
	RecordRunResult(ctx context.Context, taskID int64, status, message string, runAt time.Time) error
	AddLog(ctx context.Context, l *model.ScheduledTaskLog) error
	ListLogs(ctx context.Context, taskID int64, page, pageSize int) ([]model.ScheduledTaskLog, int64, error)
	PurgeLogsBefore(ctx context.Context, before time.Time) error
}

type ScheduledTaskService struct {
	dao       ScheduledTaskRepo
	onChanged func()
}

func NewScheduledTaskService(dao ScheduledTaskRepo) *ScheduledTaskService {
	return &ScheduledTaskService{dao: dao, onChanged: func() {}}
}

func (s *ScheduledTaskService) SetOnChanged(fn func()) {
	if fn == nil {
		s.onChanged = func() {}
		return
	}
	s.onChanged = fn
}

func (s *ScheduledTaskService) List(ctx context.Context, page, pageSize int) ([]model.ScheduledTask, int64, error) {
	return s.dao.List(ctx, page, pageSize)
}

func (s *ScheduledTaskService) Create(ctx context.Context, name, spec, command string, timeoutSec int, concurrencyPolicy string, enabled bool) (*model.ScheduledTask, error) {
	name = strings.TrimSpace(name)
	spec = strings.TrimSpace(spec)
	command = strings.TrimSpace(command)
	if name == "" || spec == "" || command == "" {
		return nil, errcode.New(errcode.BadRequest, errcode.KeyInvalidParam)
	}
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	if timeoutSec > 3600 {
		return nil, errcode.New(errcode.BadRequest, errcode.KeyInvalidParam)
	}
	concurrencyPolicy = normalizeConcurrencyPolicy(concurrencyPolicy)
	task := &model.ScheduledTask{
		Name:              name,
		Spec:              spec,
		Command:           command,
		TimeoutSec:        timeoutSec,
		ConcurrencyPolicy: concurrencyPolicy,
		Enabled:           enabled,
	}
	if err := s.dao.Create(ctx, task); err != nil {
		return nil, err
	}
	s.onChanged()
	return task, nil
}

func (s *ScheduledTaskService) Update(ctx context.Context, id int64, name, spec, command *string, timeoutSec *int, concurrencyPolicy *string, enabled *bool) (*model.ScheduledTask, error) {
	task, err := s.dao.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.BadRequest, errcode.KeyInvalidParam)
		}
		return nil, err
	}
	if name != nil {
		task.Name = strings.TrimSpace(*name)
	}
	if spec != nil {
		task.Spec = strings.TrimSpace(*spec)
	}
	if command != nil {
		task.Command = strings.TrimSpace(*command)
	}
	if enabled != nil {
		task.Enabled = *enabled
	}
	if timeoutSec != nil {
		task.TimeoutSec = *timeoutSec
	}
	if concurrencyPolicy != nil {
		task.ConcurrencyPolicy = normalizeConcurrencyPolicy(*concurrencyPolicy)
	}
	if task.Name == "" || task.Spec == "" || task.Command == "" {
		return nil, errcode.New(errcode.BadRequest, errcode.KeyInvalidParam)
	}
	if task.TimeoutSec <= 0 || task.TimeoutSec > 3600 {
		return nil, errcode.New(errcode.BadRequest, errcode.KeyInvalidParam)
	}
	task.ConcurrencyPolicy = normalizeConcurrencyPolicy(task.ConcurrencyPolicy)
	if err := s.dao.Update(ctx, task); err != nil {
		return nil, err
	}
	s.onChanged()
	return task, nil
}

func normalizeConcurrencyPolicy(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "allow":
		return "allow"
	default:
		return "forbid"
	}
}

func (s *ScheduledTaskService) Delete(ctx context.Context, id int64) error {
	if err := s.dao.Delete(ctx, id); err != nil {
		return err
	}
	s.onChanged()
	return nil
}

func (s *ScheduledTaskService) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	if _, err := s.dao.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.New(errcode.BadRequest, errcode.KeyInvalidParam)
		}
		return err
	}
	if err := s.dao.SetEnabled(ctx, id, enabled); err != nil {
		return err
	}
	s.onChanged()
	return nil
}

func (s *ScheduledTaskService) RunNow(ctx context.Context, id int64) error {
	task, err := s.dao.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.New(errcode.BadRequest, errcode.KeyInvalidParam)
		}
		return err
	}
	return s.executeTask(ctx, task)
}

func (s *ScheduledTaskService) ListLogs(ctx context.Context, taskID int64, page, pageSize int) ([]model.ScheduledTaskLog, int64, error) {
	return s.dao.ListLogs(ctx, taskID, page, pageSize)
}

func (s *ScheduledTaskService) ExecuteTaskByID(ctx context.Context, taskID int64) error {
	task, err := s.dao.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	return s.executeTask(ctx, task)
}

func (s *ScheduledTaskService) ListEnabledTasks(ctx context.Context) ([]model.ScheduledTask, error) {
	return s.dao.ListEnabled(ctx)
}

func (s *ScheduledTaskService) executeTask(ctx context.Context, task *model.ScheduledTask) error {
	start := time.Now()
	timeout := task.TimeoutSec
	if timeout <= 0 {
		timeout = 30
	}
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()
	status := "success"
	msg := "ok"
	out, execErr := runCommand(runCtx, task.Command)
	if execErr != nil {
		status = "failed"
		msg = execErr.Error()
	}
	logRow := &model.ScheduledTaskLog{
		TaskID:       task.ID,
		Status:       status,
		Output:       limitText(out, 4000),
		ErrorMessage: limitText(msg, 2000),
		StartedAt:    start,
		FinishedAt:   time.Now(),
		DurationMS:   time.Since(start).Milliseconds(),
	}
	_ = s.dao.AddLog(ctx, logRow)
	_ = s.dao.RecordRunResult(ctx, task.ID, status, limitText(msg, 255), logRow.FinishedAt)
	if execErr != nil {
		return fmt.Errorf("execute command(%s): %w", task.Command, execErr)
	}
	return nil
}

func (s *ScheduledTaskService) PurgeLogs(ctx context.Context, retentionDays int) error {
	if retentionDays <= 0 {
		return nil
	}
	before := time.Now().AddDate(0, 0, -retentionDays)
	return s.dao.PurgeLogsBefore(ctx, before)
}

func runCommand(ctx context.Context, command string) (string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func limitText(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n]
}
