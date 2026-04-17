package job

import "github.com/hibiken/asynq"

// NewInspector 创建 Asynq Inspector（用于失败任务治理）。
func NewInspector(redisAddr, password string, db int) *asynq.Inspector {
	if redisAddr == "" {
		return nil
	}
	return asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: password,
		DB:       db,
	})
}
