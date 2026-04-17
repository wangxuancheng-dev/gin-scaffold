package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"gin-scaffold/config"
	"gin-scaffold/internal/dao"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/tenant"
	"gin-scaffold/pkg/eventbus"
	"gin-scaffold/pkg/logger"
)

type OutboxDispatcher struct {
	dao *dao.OutboxDAO
	cfg config.OutboxConfig
}

func NewOutboxDispatcher(d *dao.OutboxDAO, cfg config.OutboxConfig) *OutboxDispatcher {
	return &OutboxDispatcher{dao: d, cfg: cfg}
}

func (d *OutboxDispatcher) Start() func() {
	if d == nil || d.dao == nil || !d.cfg.Enabled {
		return func() {}
	}
	interval := time.Duration(d.cfg.PollIntervalSec) * time.Second
	if interval <= 0 {
		interval = 2 * time.Second
	}
	stop := make(chan struct{})
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				d.dispatchOnce(context.Background())
			case <-stop:
				return
			}
		}
	}()
	return func() { close(stop) }
}

func (d *OutboxDispatcher) dispatchOnce(ctx context.Context) {
	batch := d.cfg.BatchSize
	if batch <= 0 {
		batch = 100
	}
	rows, err := d.dao.FetchDue(ctx, batch)
	if err != nil {
		logger.ErrorX("outbox fetch due failed", zap.Error(err))
		return
	}
	for i := range rows {
		d.handleOne(ctx, &rows[i])
	}
}

func (d *OutboxDispatcher) handleOne(ctx context.Context, row *model.OutboxEvent) {
	if row == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			d.failOne(ctx, row, fmt.Errorf("panic: %v", r))
		}
	}()
	runCtx := tenant.WithContext(ctx, row.TenantID)
	payload := map[string]any{}
	if strings.TrimSpace(row.Payload) != "" {
		if err := json.Unmarshal([]byte(row.Payload), &payload); err != nil {
			d.failOne(runCtx, row, fmt.Errorf("decode payload: %w", err))
			return
		}
	}
	eventbus.Default().Emit(runCtx, eventbus.Event{
		Name:    row.Topic,
		Payload: payload,
	})
	if err := d.dao.MarkPublished(runCtx, row.ID); err != nil {
		logger.ErrorX("outbox mark published failed", zap.Int64("id", row.ID), zap.Error(err))
	}
}

func (d *OutboxDispatcher) failOne(ctx context.Context, row *model.OutboxEvent, err error) {
	attempts := row.Attempts + 1
	backoff := d.cfg.RetryBackoffSec
	if backoff <= 0 {
		backoff = 5
	}
	if attempts >= row.MaxAttempts {
		if markErr := d.dao.MarkDead(ctx, row.ID, attempts, safeErr(err)); markErr != nil {
			logger.ErrorX("outbox mark dead failed", zap.Int64("id", row.ID), zap.Error(markErr))
		}
		return
	}
	nextRun := time.Now().Add(time.Duration(backoff*attempts) * time.Second)
	if markErr := d.dao.MarkRetry(ctx, row.ID, attempts, nextRun, safeErr(err)); markErr != nil {
		logger.ErrorX("outbox mark retry failed", zap.Int64("id", row.ID), zap.Error(markErr))
	}
}

func safeErr(err error) string {
	if err == nil {
		return ""
	}
	s := strings.TrimSpace(err.Error())
	if len(s) > 500 {
		return s[:500]
	}
	return s
}
