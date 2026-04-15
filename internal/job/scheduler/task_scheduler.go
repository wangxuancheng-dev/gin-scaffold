package scheduler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"gin-scaffold/config"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/service"
	"gin-scaffold/pkg/logger"
	appredis "gin-scaffold/pkg/redis"
)

type taskScheduler struct {
	svc           *service.ScheduledTaskService
	cron          *cron.Cron
	retentionDays int
	lockEnabled   bool
	lockTTL       time.Duration
	lockPrefix    string
	mu            sync.Mutex
	entries       map[int64]cron.EntryID
	running       map[int64]struct{}
	stopCh        chan struct{}
}

func StartTaskScheduler(svc *service.ScheduledTaskService, cfg config.SchedulerConfig) (func(), error) {
	enabled := cfg.Enabled
	withSeconds := cfg.WithSeconds
	if svc == nil || !enabled {
		return func() {}, nil
	}
	lockTTL := time.Duration(cfg.LockTTLSeconds) * time.Second
	if lockTTL <= 0 {
		lockTTL = 120 * time.Second
	}
	lockPrefix := strings.TrimSpace(cfg.LockPrefix)
	if lockPrefix == "" {
		lockPrefix = "scheduler:task:lock:"
	}
	var c *cron.Cron
	if withSeconds {
		c = cron.New(cron.WithSeconds())
	} else {
		c = cron.New()
	}
	ts := &taskScheduler{
		svc:           svc,
		cron:          c,
		retentionDays: cfg.LogRetentionDays,
		lockEnabled:   cfg.LockEnabled,
		lockTTL:       lockTTL,
		lockPrefix:    lockPrefix,
		entries:       map[int64]cron.EntryID{},
		running:       map[int64]struct{}{},
		stopCh:        make(chan struct{}),
	}
	if err := ts.sync(context.Background()); err != nil {
		return nil, err
	}
	c.Start()
	go ts.loopSync()
	logger.InfoX("db task scheduler started", zap.Bool("with_seconds", withSeconds))
	return ts.stop, nil
}

func (s *taskScheduler) loopSync() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if purgeErr := s.svc.PurgeLogs(context.Background(), s.retentionDays); purgeErr != nil {
				logger.ErrorX("purge task logs failed", zap.Error(purgeErr))
			}
			if err := s.sync(context.Background()); err != nil {
				logger.ErrorX("sync scheduled tasks failed", zap.Error(err))
			}
		case <-s.stopCh:
			return
		}
	}
}

func (s *taskScheduler) sync(ctx context.Context) error {
	rows, err := s.svc.ListEnabledTasks(ctx)
	if err != nil {
		return err
	}
	want := map[int64]model.ScheduledTask{}
	for _, r := range rows {
		want[r.ID] = r
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for taskID, entry := range s.entries {
		if _, ok := want[taskID]; !ok {
			s.cron.Remove(entry)
			delete(s.entries, taskID)
		}
	}
	for _, t := range rows {
		if _, ok := s.entries[t.ID]; ok {
			continue
		}
		taskID := t.ID
		spec := t.Spec
		entryID, err := s.cron.AddFunc(spec, func() {
			if runErr := s.executeWithGuards(taskID); runErr != nil {
				logger.Channel("daily", "task_scheduler.log").Error("execute scheduled task failed",
					zap.Int64("task_id", taskID),
					zap.Error(runErr),
				)
			}
		})
		if err != nil {
			return fmt.Errorf("register task id=%d spec=%s: %w", taskID, spec, err)
		}
		s.entries[t.ID] = entryID
	}
	return nil
}

func (s *taskScheduler) stop() {
	close(s.stopCh)
	ctx := s.cron.Stop()
	<-ctx.Done()
	logger.InfoX("db task scheduler stopped")
}

func (s *taskScheduler) executeWithGuards(taskID int64) error {
	if !s.enterLocal(taskID) {
		return nil
	}
	defer s.leaveLocal(taskID)

	unlock := func() {}
	if s.lockEnabled {
		u, ok, err := s.acquireDistributedLock(context.Background(), taskID)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		unlock = u
	}
	defer unlock()

	return s.svc.ExecuteTaskByID(context.Background(), taskID)
}

func (s *taskScheduler) enterLocal(taskID int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.running[taskID]; ok {
		return false
	}
	s.running[taskID] = struct{}{}
	return true
}

func (s *taskScheduler) leaveLocal(taskID int64) {
	s.mu.Lock()
	delete(s.running, taskID)
	s.mu.Unlock()
}

func (s *taskScheduler) acquireDistributedLock(ctx context.Context, taskID int64) (func(), bool, error) {
	rc := appredis.Client()
	if rc == nil {
		return nil, false, fmt.Errorf("scheduler lock enabled but redis client is nil")
	}
	key := fmt.Sprintf("%s%d", s.lockPrefix, taskID)
	token := uuid.NewString()
	ok, err := rc.SetNX(ctx, key, token, s.lockTTL).Result()
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	unlock := func() {
		// Compare-and-delete to avoid releasing others' lock.
		_, _ = rc.Eval(ctx, `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
end
return 0
`, []string{key}, token).Result()
	}
	return unlock, true, nil
}
