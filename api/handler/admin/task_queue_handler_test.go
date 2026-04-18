package adminhandler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"

	"gin-scaffold/internal/pkg/errcode"
)

type stubQueueInspector struct {
	queueInfo      map[string]*asynq.QueueInfo
	queueInfoErr   map[string]error
	retryTasks     []*asynq.TaskInfo
	retryErr       error
	archivedTasks  []*asynq.TaskInfo
	archivedErr    error
	runTaskErr     error
	archiveTaskErr error
}

func (s *stubQueueInspector) ListRetryTasks(queue string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
	return s.retryTasks, s.retryErr
}

func (s *stubQueueInspector) ListArchivedTasks(queue string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
	return s.archivedTasks, s.archivedErr
}

func (s *stubQueueInspector) GetQueueInfo(queue string) (*asynq.QueueInfo, error) {
	if s.queueInfoErr != nil {
		if e, ok := s.queueInfoErr[queue]; ok && e != nil {
			return nil, e
		}
	}
	if s.queueInfo != nil {
		if info, ok := s.queueInfo[queue]; ok {
			return info, nil
		}
	}
	return &asynq.QueueInfo{}, nil
}

func (s *stubQueueInspector) RunTask(queue, id string) error {
	return s.runTaskErr
}

func (s *stubQueueInspector) ArchiveTask(queue, id string) error {
	return s.archiveTaskErr
}

var _ queueInspector = (*stubQueueInspector)(nil)

func TestTaskQueueHandler_Summary_nilInspector(t *testing.T) {
	h := NewTaskQueueHandler(nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/task-queues/summary", nil)
	h.Summary(c)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestTaskQueueHandler_Summary_mixedQueues(t *testing.T) {
	ins := &stubQueueInspector{
		queueInfo: map[string]*asynq.QueueInfo{
			"default": {Pending: 1, Active: 2},
		},
		queueInfoErr: map[string]error{
			"critical": errors.New("boom"),
		},
	}
	h := NewTaskQueueHandler(ins)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/task-queues/summary", nil)
	h.Summary(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestTaskQueueHandler_FailedList_retryAndArchived(t *testing.T) {
	ins := &stubQueueInspector{
		retryTasks: []*asynq.TaskInfo{
			{ID: "t1", Type: "x", Queue: "default", State: asynq.TaskStateRetry, Payload: []byte("{}")},
		},
		archivedTasks: []*asynq.TaskInfo{
			{ID: "t2", Type: "y", Queue: "default", State: asynq.TaskStateArchived},
		},
	}
	h := NewTaskQueueHandler(ins)

	t.Run("retry", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/task-queues/failed?state=retry&queue=default", nil)
		h.FailedList(c)
		if w.Code != http.StatusOK {
			t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("archived", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/task-queues/failed?state=archived", nil)
		h.FailedList(c)
		if w.Code != http.StatusOK {
			t.Fatalf("code=%d", w.Code)
		}
	})
}

func TestTaskQueueHandler_Retry_Archive_ok(t *testing.T) {
	ins := &stubQueueInspector{}
	h := NewTaskQueueHandler(ins)

	t.Run("retry", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "queue", Value: "default"}, {Key: "task_id", Value: "abc"}}
		c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/admin/task-queues/default/failed/abc/retry", nil)
		h.Retry(c)
		if w.Code != http.StatusOK {
			t.Fatalf("code=%d", w.Code)
		}
	})

	t.Run("archive", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "queue", Value: "default"}, {Key: "task_id", Value: "abc"}}
		c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/admin/task-queues/default/failed/abc/archive", nil)
		h.Archive(c)
		if w.Code != http.StatusOK {
			t.Fatalf("code=%d", w.Code)
		}
	})
}

func TestTaskQueueHandler_Retry_bizError(t *testing.T) {
	ins := &stubQueueInspector{runTaskErr: errcode.New(errcode.BadRequest, errcode.KeyInvalidParam)}
	h := NewTaskQueueHandler(ins)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "queue", Value: "default"}, {Key: "task_id", Value: "x"}}
	c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/admin/task-queues/default/failed/x/retry", nil)
	h.Retry(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", w.Code)
	}
}
