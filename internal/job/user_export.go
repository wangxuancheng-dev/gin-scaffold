package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	appredis "gin-scaffold/pkg/redis"
)

const (
	userExportStatusPrefix = "user:export:status:"
	userExportStatusTTL    = 7 * 24 * time.Hour
)

// UserExportPayload 用户异步导出任务载荷。
type UserExportPayload struct {
	TaskID    string   `json:"task_id"`
	Operator  int64    `json:"operator"`
	Username  string   `json:"username,omitempty"`
	Nickname  string   `json:"nickname,omitempty"`
	Fields    []string `json:"fields"`
	WithRole  bool     `json:"with_role"`
	FileType  string   `json:"file_type"` // csv
}

// UserExportStatus 用户导出任务状态。
type UserExportStatus struct {
	TaskID       string `json:"task_id"`
	State        string `json:"state"` // queued|running|success|failed
	ProgressRows int64  `json:"progress_rows"`
	FileKey      string `json:"file_key,omitempty"`
	Error        string `json:"error,omitempty"`
	Filter       string `json:"filter,omitempty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func userExportStatusKey(taskID string) string {
	return userExportStatusPrefix + taskID
}

func SetUserExportStatus(ctx context.Context, st *UserExportStatus) error {
	if st == nil || st.TaskID == "" {
		return fmt.Errorf("user export status invalid")
	}
	if st.CreatedAt == "" {
		st.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	st.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	b, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return appredis.Set(ctx, userExportStatusKey(st.TaskID), string(b), userExportStatusTTL)
}

func GetUserExportStatus(ctx context.Context, taskID string) (*UserExportStatus, error) {
	raw, err := appredis.Get(ctx, userExportStatusKey(taskID))
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var st UserExportStatus
	if err := json.Unmarshal([]byte(raw), &st); err != nil {
		return nil, err
	}
	return &st, nil
}
