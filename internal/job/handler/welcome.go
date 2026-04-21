package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"gin-scaffold/internal/job"
	"gin-scaffold/pkg/logger"
	"gin-scaffold/pkg/metrics"
)

// WelcomeHandler 处理欢迎任务。
type WelcomeHandler struct{}

// ProcessTask 实现 asynq.Handler。
func (WelcomeHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	start := time.Now()
	status := "ok"
	defer func() {
		metrics.ObserveAsynqTask(job.TypeWelcomeEmail, status, start)
	}()

	var p job.WelcomePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		status = "bad_payload"
		return err
	}
	logger.InfoX("welcome task", zap.Int64("uid", p.UserID), zap.String("user", p.Username))
	return nil
}
