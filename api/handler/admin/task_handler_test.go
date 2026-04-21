package adminhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	adminreq "gin-scaffold/api/request/admin"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/internal/service/port"
)

type stubScheduledTaskService struct {
	listRows []model.ScheduledTask
	listTot  int64
	listErr  error

	createRow *model.ScheduledTask
	createErr error

	updateRow *model.ScheduledTask
	updateErr error

	deleteErr error

	setEnabledErr error
	runNowErr     error

	logRows []model.ScheduledTaskLog
	logTot  int64
	logErr  error
}

func (s *stubScheduledTaskService) List(ctx context.Context, page, pageSize int) ([]model.ScheduledTask, int64, error) {
	return s.listRows, s.listTot, s.listErr
}

func (s *stubScheduledTaskService) Create(ctx context.Context, name, spec, command string, timeoutSec int, concurrencyPolicy string, enabled bool) (*model.ScheduledTask, error) {
	if s.createRow != nil {
		cp := *s.createRow
		return &cp, s.createErr
	}
	return nil, s.createErr
}

func (s *stubScheduledTaskService) Update(ctx context.Context, id int64, name, spec, command *string, timeoutSec *int, concurrencyPolicy *string, enabled *bool) (*model.ScheduledTask, error) {
	if s.updateRow != nil {
		cp := *s.updateRow
		return &cp, s.updateErr
	}
	return nil, s.updateErr
}

func (s *stubScheduledTaskService) Delete(ctx context.Context, id int64) error {
	return s.deleteErr
}

func (s *stubScheduledTaskService) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	return s.setEnabledErr
}

func (s *stubScheduledTaskService) RunNow(ctx context.Context, id int64) error {
	return s.runNowErr
}

func (s *stubScheduledTaskService) ListLogs(ctx context.Context, taskID int64, page, pageSize int) ([]model.ScheduledTaskLog, int64, error) {
	return s.logRows, s.logTot, s.logErr
}

var _ port.ScheduledTaskService = (*stubScheduledTaskService)(nil)

func TestTaskHandler_List_ok(t *testing.T) {
	svc := &stubScheduledTaskService{
		listRows: []model.ScheduledTask{{ID: 1, Name: "t1", Spec: "@daily", Command: "ping", Enabled: true}},
		listTot:  1,
	}
	h := NewTaskHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/tasks?page=1&page_size=10", nil)
	h.List(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestTaskHandler_Create_ok(t *testing.T) {
	now := time.Now()
	svc := &stubScheduledTaskService{
		createRow: &model.ScheduledTask{ID: 2, Name: "job", Spec: "0 * * * *", Command: "echo hi", CreatedAt: now, UpdatedAt: now},
	}
	h := NewTaskHandler(svc)
	body, _ := json.Marshal(adminreq.TaskCreateRequest{
		Name: "job", Spec: "0 * * * *", Command: "echo hi",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/admin/tasks", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Create(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestTaskHandler_Create_bizError(t *testing.T) {
	svc := &stubScheduledTaskService{createErr: errcode.New(errcode.Conflict, errcode.KeyTaskAlreadyRunning)}
	h := NewTaskHandler(svc)
	body, _ := json.Marshal(adminreq.TaskCreateRequest{Name: "job", Spec: "0 * * * *", Command: "echo hi"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/admin/tasks", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Create(c)
	if w.Code != http.StatusConflict {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestTaskHandler_NotFoundErrorMappedTo404(t *testing.T) {
	svc := &stubScheduledTaskService{runNowErr: errcode.New(errcode.NotFound, errcode.KeyNotFound)}
	h := NewTaskHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "3"}}
	c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/admin/tasks/3/run", nil)
	h.RunNow(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestTaskHandler_Update_Delete_Toggle_RunNow_Logs(t *testing.T) {
	now := time.Now()
	row := &model.ScheduledTask{ID: 3, Name: "x", Spec: "0 * * * *", Command: "c", CreatedAt: now, UpdatedAt: now}
	svc := &stubScheduledTaskService{
		updateRow: row,
		logRows:   []model.ScheduledTaskLog{{ID: 1, TaskID: 3, Status: "ok"}},
		logTot:    1,
	}
	h := NewTaskHandler(svc)

	t.Run("update", func(t *testing.T) {
		n := "newname"
		body, _ := json.Marshal(adminreq.TaskUpdateRequest{Name: &n})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "3"}}
		c.Request = httptest.NewRequest(http.MethodPut, "http://localhost/admin/tasks/3", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Update(c)
		if w.Code != http.StatusOK {
			t.Fatalf("code=%d", w.Code)
		}
	})

	t.Run("delete", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "3"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "http://localhost/admin/tasks/3", nil)
		h.Delete(c)
		if w.Code != http.StatusOK {
			t.Fatalf("code=%d", w.Code)
		}
	})

	t.Run("toggle", func(t *testing.T) {
		body, _ := json.Marshal(adminreq.TaskToggleRequest{Enabled: false})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "3"}}
		c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/admin/tasks/3/toggle", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Toggle(c)
		if w.Code != http.StatusOK {
			t.Fatalf("code=%d", w.Code)
		}
	})

	t.Run("run_now", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "3"}}
		c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/admin/tasks/3/run", nil)
		h.RunNow(c)
		if w.Code != http.StatusOK {
			t.Fatalf("code=%d", w.Code)
		}
	})

	t.Run("logs", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "3"}}
		c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/tasks/3/logs?page=1&page_size=5", nil)
		h.Logs(c)
		if w.Code != http.StatusOK {
			t.Fatalf("code=%d", w.Code)
		}
	})
}

func TestTaskHandler_List_serviceError(t *testing.T) {
	h := NewTaskHandler(&stubScheduledTaskService{listErr: errors.New("db")})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/tasks", nil)
	h.List(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d", w.Code)
	}
}
