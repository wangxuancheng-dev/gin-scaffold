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
	return c.EnqueueTaskInQueue(ctx, c.queue, taskType, payload, uniqueWindowSec)
}

// EnqueueTaskInQueue 投递普通任务（可指定队列，可选去重窗口）。
func (c *Client) EnqueueTaskInQueue(ctx context.Context, queueName string, taskType string, payload any, uniqueWindowSec int) error {
	return c.enqueue(ctx, queueName, taskType, payload, uniqueWindowSec)
}

// EnqueueTaskAfter 在默认队列投递任务，延迟 delay 后再由 worker 消费（delay<=0 时等价于立即入队）。
// 适用于订单超时关单、延迟重试通知等「相对当前时间」的一次性调度。
func (c *Client) EnqueueTaskAfter(ctx context.Context, taskType string, payload any, uniqueWindowSec int, delay time.Duration) error {
	return c.EnqueueTaskInQueueAfter(ctx, c.queue, taskType, payload, uniqueWindowSec, delay)
}

// EnqueueTaskInQueueAfter 在指定队列投递任务，延迟 delay 后再消费。
func (c *Client) EnqueueTaskInQueueAfter(ctx context.Context, queueName string, taskType string, payload any, uniqueWindowSec int, delay time.Duration) error {
	if delay <= 0 {
		return c.enqueue(ctx, queueName, taskType, payload, uniqueWindowSec)
	}
	return c.enqueue(ctx, queueName, taskType, payload, uniqueWindowSec, asynq.ProcessIn(delay))
}

// EnqueueTaskAt 在默认队列投递任务，在 at 时刻（及之后首次轮询）由 worker 消费；at 为零值时等价于立即入队。
func (c *Client) EnqueueTaskAt(ctx context.Context, taskType string, payload any, uniqueWindowSec int, at time.Time) error {
	return c.EnqueueTaskInQueueAt(ctx, c.queue, taskType, payload, uniqueWindowSec, at)
}

// EnqueueTaskInQueueAt 在指定队列投递任务，在绝对时间 at 执行。
func (c *Client) EnqueueTaskInQueueAt(ctx context.Context, queueName string, taskType string, payload any, uniqueWindowSec int, at time.Time) error {
	if at.IsZero() {
		return c.enqueue(ctx, queueName, taskType, payload, uniqueWindowSec)
	}
	return c.enqueue(ctx, queueName, taskType, payload, uniqueWindowSec, asynq.ProcessAt(at))
}

// EnqueueUniqueAfter 使用配置的去重窗口投递任务，并延迟 delay 后执行。
func (c *Client) EnqueueUniqueAfter(ctx context.Context, taskType string, payload any, delay time.Duration) error {
	return c.EnqueueUniqueInQueueAfter(ctx, c.queue, taskType, payload, delay)
}

// EnqueueUniqueInQueueAfter 在指定队列使用去重窗口投递，并延迟 delay 后执行。
func (c *Client) EnqueueUniqueInQueueAfter(ctx context.Context, queueName string, taskType string, payload any, delay time.Duration) error {
	return c.EnqueueTaskInQueueAfter(ctx, queueName, taskType, payload, c.dedupWindowSec, delay)
}

// EnqueueUniqueAt 使用配置的去重窗口投递任务，在 at 时刻执行。
func (c *Client) EnqueueUniqueAt(ctx context.Context, taskType string, payload any, at time.Time) error {
	return c.EnqueueUniqueInQueueAt(ctx, c.queue, taskType, payload, at)
}

// EnqueueUniqueInQueueAt 在指定队列使用去重窗口投递，在 at 时刻执行。
func (c *Client) EnqueueUniqueInQueueAt(ctx context.Context, queueName string, taskType string, payload any, at time.Time) error {
	return c.EnqueueTaskInQueueAt(ctx, queueName, taskType, payload, c.dedupWindowSec, at)
}

func (c *Client) enqueue(ctx context.Context, queueName string, taskType string, payload any, uniqueWindowSec int, extra ...asynq.Option) error {
	if c == nil || c.c == nil {
		return nil
	}
	if queueName == "" {
		queueName = c.queue
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskType, b)
	opts := []asynq.Option{
		asynq.MaxRetry(c.maxRetry),
		asynq.Timeout(time.Duration(c.timeoutSec) * time.Second),
		asynq.Queue(queueName),
	}
	if uniqueWindowSec > 0 {
		// asynq.Unique(ttl) 要求 ttl >= 1s，这里做兜底。
		ttl := uniqueWindowSec
		if ttl < 1 {
			ttl = 1
		}
		opts = append(opts, asynq.Unique(time.Duration(ttl)*time.Second))
	}
	opts = append(opts, extra...)
	_, err = c.c.EnqueueContext(ctx, task, opts...)
	return err
}

// EnqueueUnique 使用配置的去重窗口投递任务。
// 去重粒度由 taskType + payload 决定，适合“重复点击提交同一业务请求”场景。
func (c *Client) EnqueueUnique(ctx context.Context, taskType string, payload any) error {
	return c.EnqueueUniqueInQueue(ctx, c.queue, taskType, payload)
}

// EnqueueUniqueInQueue 使用配置的去重窗口投递任务（可指定队列）。
func (c *Client) EnqueueUniqueInQueue(ctx context.Context, queueName string, taskType string, payload any) error {
	return c.EnqueueTaskInQueue(ctx, queueName, taskType, payload, c.dedupWindowSec)
}
