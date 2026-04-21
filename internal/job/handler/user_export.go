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
	"gin-scaffold/internal/model"
	"gin-scaffold/pkg/db"
	"gin-scaffold/pkg/metrics"
	"gin-scaffold/pkg/storage"
)

// UserExportHandler 处理用户异步导出。
type UserExportHandler struct{}

func (UserExportHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	start := time.Now()
	status := "ok"
	defer func() {
		metrics.ObserveAsynqTask(job.TypeUserExport, status, start)
	}()

	fail := func(taskID string, err error) error {
		status = "failed"
		return markUserExportFailed(ctx, taskID, err)
	}

	var p job.UserExportPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		status = "bad_payload"
		return err
	}
	now := time.Now().UTC()
	filter := buildUserExportFilter(p)
	_ = job.SetUserExportStatus(ctx, &job.UserExportStatus{
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
	userDAO := dao.NewUserDAO(gdb)
	tmpFile := filepath.Join(os.TempDir(), "user-export-"+p.TaskID+".csv")
	f, err := os.Create(tmpFile)
	if err != nil {
		return fail(p.TaskID, err)
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(tmpFile)
	}()

	fields := sanitizeUserFields(p.Fields)
	withRole := p.WithRole
	w := csv.NewWriter(f)
	_ = w.Write(fields)
	var (
		lastID int64
		total  int64
	)
	query := model.UserQuery{Username: p.Username, Nickname: p.Nickname}
	for {
		rows, qErr := userDAO.ListAfterID(ctx, query, lastID, 1000)
		if qErr != nil {
			return fail(p.TaskID, qErr)
		}
		if len(rows) == 0 {
			break
		}
		roleMap := map[int64]string{}
		if withRole {
			ids := make([]int64, 0, len(rows))
			for _, u := range rows {
				ids = append(ids, u.ID)
			}
			roleMap, qErr = userDAO.GetPrimaryRoles(ctx, ids)
			if qErr != nil {
				return fail(p.TaskID, qErr)
			}
		}
		for _, u := range rows {
			record := make([]string, 0, len(fields))
			for _, field := range fields {
				switch field {
				case "id":
					record = append(record, strconv.FormatInt(u.ID, 10))
				case "username":
					record = append(record, u.Username)
				case "nickname":
					record = append(record, u.Nickname)
				case "created_at":
					record = append(record, u.CreatedAt.Format(time.RFC3339))
				case "role":
					role := roleMap[u.ID]
					if role == "" {
						role = "user"
					}
					record = append(record, role)
				}
			}
			_ = w.Write(record)
			total++
			lastID = u.ID
		}
		w.Flush()
		_ = job.SetUserExportStatus(ctx, &job.UserExportStatus{
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
	key := "exports/users/" + time.Now().UTC().Format("20060102") + "/" + uuid.NewString() + ".csv"
	if pc, ok := sp.(storage.PutContentTyper); ok {
		if err := pc.PutContentType(ctx, key, "text/csv", f); err != nil {
			return fail(p.TaskID, err)
		}
	} else if err := sp.Put(ctx, key, f); err != nil {
		return fail(p.TaskID, err)
	}
	return job.SetUserExportStatus(ctx, &job.UserExportStatus{
		TaskID:       p.TaskID,
		State:        "success",
		ProgressRows: total,
		FileKey:      key,
		Filter:       filter,
		CreatedAt:    now.Format(time.RFC3339),
	})
}

func sanitizeUserFields(fields []string) []string {
	allowed := map[string]struct{}{"id": {}, "username": {}, "nickname": {}, "created_at": {}, "role": {}}
	if len(fields) == 0 {
		return []string{"id", "username", "nickname", "created_at"}
	}
	out := make([]string, 0, len(fields))
	seen := map[string]struct{}{}
	for _, f := range fields {
		k := strings.ToLower(strings.TrimSpace(f))
		if _, ok := allowed[k]; !ok {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	if len(out) == 0 {
		return []string{"id", "username", "nickname", "created_at"}
	}
	return out
}

func buildUserExportFilter(p job.UserExportPayload) string {
	parts := []string{"file_type=csv"}
	if s := strings.TrimSpace(p.Username); s != "" {
		parts = append(parts, "username~="+s)
	}
	if s := strings.TrimSpace(p.Nickname); s != "" {
		parts = append(parts, "nickname~="+s)
	}
	if len(p.Fields) > 0 {
		parts = append(parts, "fields="+strings.Join(sanitizeUserFields(p.Fields), ","))
	}
	if p.WithRole {
		parts = append(parts, "with_role=true")
	}
	out := strings.Join(parts, "&")
	if len(out) > 1024 {
		return out[:1024]
	}
	return out
}

func markUserExportFailed(ctx context.Context, taskID string, err error) error {
	msg := ""
	if err != nil {
		msg = strings.TrimSpace(err.Error())
	}
	_ = job.SetUserExportStatus(ctx, &job.UserExportStatus{
		TaskID:    taskID,
		State:     "failed",
		Error:     msg,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	return err
}
