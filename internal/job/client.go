package job

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

// WelcomePayload 欢迎任务载荷。
type WelcomePayload struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

// Client Asynq 客户端封装。
type Client struct {
	c              *asynq.Client
	queue          string
	maxRetry       int
	timeoutSec     int
	dedupWindowSec int
}

// NewClient 创建客户端。
func NewClient(redisAddr, password string, db int, queue string, maxRetry, timeoutSec, dedupWindowSec int) *Client {
	if queue == "" {
		queue = "default"
	}
	if maxRetry < 0 {
		maxRetry = 0
	}
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	if dedupWindowSec < 0 {
		dedupWindowSec = 0
	}
	return &Client{
		c: asynq.NewClient(asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: password,
			DB:       db,
		}),
		queue:          queue,
		maxRetry:       maxRetry,
		timeoutSec:     timeoutSec,
		dedupWindowSec: dedupWindowSec,
	}
}

// Close 关闭连接。
func (c *Client) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	return c.c.Close()
}

// EnqueueWelcome 投递欢迎任务（带重试与超时元数据）。
func (c *Client) EnqueueWelcome(ctx context.Context, userID int64, username string) error {
	p := WelcomePayload{UserID: userID, Username: username}
	// Welcome 消息默认不做去重，避免漏发；资金等关键业务应使用 EnqueueUnique。
	return c.EnqueueTask(ctx, TypeWelcomeEmail, p, 0)
}

// EnqueueTask 投递普通任务（不去重）。
func (c *Client) EnqueueTask(ctx context.Context, taskType string, payload any, uniqueWindowSec int) error {
	if c == nil || c.c == nil {
		return nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskType, b)
	opts := []asynq.Option{
		asynq.MaxRetry(c.maxRetry),
		asynq.Timeout(time.Duration(c.timeoutSec) * time.Second),
		asynq.Queue(c.queue),
	}
	if uniqueWindowSec > 0 {
		opts = append(opts, asynq.Unique(time.Duration(uniqueWindowSec)*time.Second))
	}
	_, err = c.c.EnqueueContext(ctx, task, opts...)
	return err
}

// EnqueueUnique 使用配置的去重窗口投递任务。
// 去重粒度由 taskType + payload 决定，适合“重复点击提交同一业务请求”场景。
func (c *Client) EnqueueUnique(ctx context.Context, taskType string, payload any) error {
	return c.EnqueueTask(ctx, taskType, payload, c.dedupWindowSec)
}
