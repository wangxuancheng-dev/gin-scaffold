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
	c *asynq.Client
}

// NewClient 创建客户端。
func NewClient(redisAddr, password string, db int) *Client {
	return &Client{
		c: asynq.NewClient(asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: password,
			DB:       db,
		}),
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
	if c == nil || c.c == nil {
		return nil
	}
	p := WelcomePayload{UserID: userID, Username: username}
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeWelcomeEmail, b)
	opts := []asynq.Option{
		asynq.MaxRetry(5),
		asynq.Timeout(30 * time.Second),
		asynq.Queue("default"),
	}
	_, err = c.c.EnqueueContext(ctx, task, opts...)
	return err
}
