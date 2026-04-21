package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"gin-scaffold/internal/dao"
	"gin-scaffold/internal/job"
	"gin-scaffold/internal/pkg/timefmt"
	"gin-scaffold/pkg/db"
	"gin-scaffold/pkg/metrics"
	"gin-scaffold/pkg/storage"
)

// AuditExportHandler 审计日志异步导出。
type AuditExportHandler struct{}

// ProcessTask 实现 asynq.Handler。
func (AuditExportHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	start := time.Now()
	status := "ok"
	defer func() {
		metrics.ObserveAsynqTask(job.TypeAuditExport, status, start)
	}()

	fail := func(taskID string, err error) error {
		status = "failed"
		return markExportFailed(ctx, taskID, err)
	}

	var p job.AuditExportPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		status = "bad_payload"
		return err
	}
	now := time.Now().UTC()
	filter := buildPayloadFilterSummary(p)
	_ = job.SetAuditExportStatus(ctx, &job.AuditExportStatus{
		TaskID:    p.TaskID,
		State:     "running",
		Filter:    filter,
		CreatedAt: now.Format(time.RFC3339),
	})

	gdb := db.DB()
	if gdb == nil {
		return fail(p.TaskID, fmt.Errorf("db not initialized"))
	}
	sp, err := storage.Require()
	if err != nil {
		return fail(p.TaskID, err)
	}
	from, err := timefmt.ParseRFC3339(p.From)
	if err != nil {
		return fail(p.TaskID, fmt.Errorf("invalid from"))
	}
	to, err := timefmt.ParseRFC3339(p.To)
	if err != nil {
		return fail(p.TaskID, fmt.Errorf("invalid to"))
	}
	q := dao.AuditLogListQuery{
		UserID:    p.UserID,
		Action:    p.Action,
		Status:    p.Status,
		PathLike:  p.PathLike,
		RequestID: p.RequestID,
		From:      &from,
		To:        &to,
	}

	daoObj := dao.NewAuditLogDAO(gdb)
	tmpFile := filepath.Join(os.TempDir(), "audit-export-"+p.TaskID+".csv")
	f, err := os.Create(tmpFile)
	if err != nil {
		return fail(p.TaskID, err)
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(tmpFile)
	}()

	w := csv.NewWriter(f)
	_ = w.Write([]string{"id", "request_id", "user_id", "role", "actor_type", "action", "path", "query", "status", "latency_ms", "client_ip", "created_at"})
	var (
		lastID int64
		total  int64
	)
	for {
		rows, qErr := daoObj.ListExportChunk(ctx, q, lastID, 1000)
		if qErr != nil {
			return fail(p.TaskID, qErr)
		}
		if len(rows) == 0 {
			break
		}
		for _, r := range rows {
			_ = w.Write([]string{
				strconv.FormatInt(r.ID, 10), r.RequestID, strconv.FormatInt(r.UserID, 10), r.Role, r.ActorType,
				r.Action, r.Path, r.Query, strconv.Itoa(r.Status), strconv.Itoa(r.LatencyMS), r.ClientIP, r.CreatedAt.Format(time.RFC3339),
			})
			lastID = r.ID
			total++
		}
		w.Flush()
		_ = job.SetAuditExportStatus(ctx, &job.AuditExportStatus{
			TaskID:       p.TaskID,
			State:        "running",
			ProgressRows: total,
			Filter:       filter,
			CreatedAt:    now.Format(time.RFC3339),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fail(p.TaskID, err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return fail(p.TaskID, err)
	}
	key := "exports/audit/" + time.Now().UTC().Format("20060102") + "/" + uuid.NewString() + ".csv"
	if pc, ok := sp.(storage.PutContentTyper); ok {
		if err := pc.PutContentType(ctx, key, "text/csv", f); err != nil {
			return fail(p.TaskID, err)
		}
	} else if err := sp.Put(ctx, key, f); err != nil {
		return fail(p.TaskID, err)
	}
	return job.SetAuditExportStatus(ctx, &job.AuditExportStatus{
		TaskID:       p.TaskID,
		State:        "success",
		ProgressRows: total,
		FileKey:      key,
		Filter:       filter,
		CreatedAt:    now.Format(time.RFC3339),
	})
}

func markExportFailed(ctx context.Context, taskID string, err error) error {
	msg := ""
	if err != nil {
		msg = strings.TrimSpace(err.Error())
	}
	_ = job.SetAuditExportStatus(ctx, &job.AuditExportStatus{
		TaskID:    taskID,
		State:     "failed",
		Error:     msg,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	return err
}

func buildPayloadFilterSummary(p job.AuditExportPayload) string {
	parts := make([]string, 0, 8)
	if p.UserID > 0 {
		parts = append(parts, "user_id="+strconv.FormatInt(p.UserID, 10))
	}
	if s := strings.TrimSpace(p.Action); s != "" {
		parts = append(parts, "action="+strings.ToUpper(s))
	}
	if p.Status > 0 {
		parts = append(parts, "status="+strconv.Itoa(p.Status))
	}
	if s := strings.TrimSpace(p.PathLike); s != "" {
		parts = append(parts, "path~="+s)
	}
	if s := strings.TrimSpace(p.RequestID); s != "" {
		parts = append(parts, "request_id="+s)
	}
	if s := strings.TrimSpace(p.From); s != "" {
		parts = append(parts, "from="+s)
	}
	if s := strings.TrimSpace(p.To); s != "" {
		parts = append(parts, "to="+s)
	}
	out := strings.Join(parts, "&")
	if len(out) > 1024 {
		return out[:1024]
	}
	return out
}
