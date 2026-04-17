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
	auditExportStatusPrefix = "audit:export:status:"
	auditExportStatusTTL    = 7 * 24 * time.Hour
)

// AuditExportPayload 审计导出异步任务载荷。
type AuditExportPayload struct {
	TaskID    string `json:"task_id"`
	Operator  int64  `json:"operator"`
	UserID    int64  `json:"user_id,omitempty"`
	From      string `json:"from"` // RFC3339
	To        string `json:"to"`   // RFC3339
	Action    string `json:"action,omitempty"`
	Status    int    `json:"status,omitempty"`
	PathLike  string `json:"path,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// AuditExportStatus 审计导出任务状态。
type AuditExportStatus struct {
	TaskID       string `json:"task_id"`
	State        string `json:"state"` // queued|running|success|failed
	ProgressRows int64  `json:"progress_rows"`
	FileKey      string `json:"file_key,omitempty"`
	Error        string `json:"error,omitempty"`
	Filter       string `json:"filter,omitempty"` // 导出筛选摘要，便于任务列表展示
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func auditExportStatusKey(taskID string) string {
	return auditExportStatusPrefix + taskID
}

// SetAuditExportStatus 保存任务状态。
func SetAuditExportStatus(ctx context.Context, st *AuditExportStatus) error {
	if st == nil || st.TaskID == "" {
		return fmt.Errorf("audit export status invalid")
	}
	if st.CreatedAt == "" {
		st.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	st.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	b, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return appredis.Set(ctx, auditExportStatusKey(st.TaskID), string(b), auditExportStatusTTL)
}

// GetAuditExportStatus 读取任务状态。
func GetAuditExportStatus(ctx context.Context, taskID string) (*AuditExportStatus, error) {
	raw, err := appredis.Get(ctx, auditExportStatusKey(taskID))
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var st AuditExportStatus
	if err := json.Unmarshal([]byte(raw), &st); err != nil {
		return nil, err
	}
	return &st, nil
}
