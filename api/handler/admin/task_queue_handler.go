package adminhandler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"

	"gin-scaffold/api/handler"
	adminreq "gin-scaffold/api/request/admin"
	"gin-scaffold/api/response"
)

type queueInspector interface {
	ListRetryTasks(queue string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)
	ListArchivedTasks(queue string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)
	RunTask(queue, id string) error
	ArchiveTask(queue, id string) error
}

type TaskQueueHandler struct {
	inspector queueInspector
}

func NewTaskQueueHandler(i queueInspector) *TaskQueueHandler {
	return &TaskQueueHandler{inspector: i}
}

// FailedList 查询失败任务（retry/archived）。
// @Summary 查询失败任务队列（后台）
// @Tags admin-task
// @Produce json
// @Param queue query string false "队列名，默认 default"
// @Param state query string false "retry(默认) 或 archived"
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页条数，默认 20，最大 100"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/task-queues/failed [get]
func (h *TaskQueueHandler) FailedList(c *gin.Context) {
	if h == nil || h.inspector == nil {
		handler.FailServiceUnavailable(c, nil, "asynq inspector unavailable")
		return
	}
	var q adminreq.QueueTaskListQuery
	_ = c.ShouldBindQuery(&q)
	queue := strings.TrimSpace(q.Queue)
	if queue == "" {
		queue = "default"
	}
	state := strings.ToLower(strings.TrimSpace(q.State))
	if state == "" {
		state = "retry"
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	opts := []asynq.ListOption{
		asynq.Page(page - 1),
		asynq.PageSize(pageSize),
	}
	var (
		list []*asynq.TaskInfo
		err  error
	)
	if state == "archived" {
		list, err = h.inspector.ListArchivedTasks(queue, opts...)
	} else {
		state = "retry"
		list, err = h.inspector.ListRetryTasks(queue, opts...)
	}
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	rows := make([]gin.H, 0, len(list))
	for _, t := range list {
		if t == nil {
			continue
		}
		rows = append(rows, gin.H{
			"id":           t.ID,
			"type":         t.Type,
			"queue":        t.Queue,
			"state":        t.State.String(),
			"max_retry":    t.MaxRetry,
			"retried":      t.Retried,
			"last_err":     t.LastErr,
			"payload":      string(t.Payload),
			"next_process": formatQueueTaskTime(t.NextProcessAt),
		})
	}
	response.OK(c, gin.H{
		"queue":     queue,
		"state":     state,
		"page":      page,
		"page_size": pageSize,
		"list":      rows,
	})
}

// Retry 立即重试失败任务。
// @Summary 重试失败任务（后台）
// @Tags admin-task
// @Produce json
// @Param queue path string true "队列名"
// @Param task_id path string true "任务ID"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/task-queues/{queue}/failed/{task_id}/retry [post]
func (h *TaskQueueHandler) Retry(c *gin.Context) {
	if h == nil || h.inspector == nil {
		handler.FailServiceUnavailable(c, nil, "asynq inspector unavailable")
		return
	}
	var uri adminreq.QueueTaskActionURI
	if err := c.ShouldBindUri(&uri); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	if err := h.inspector.RunTask(strings.TrimSpace(uri.Queue), strings.TrimSpace(uri.TaskID)); err != nil {
		handler.FailByError(c, err, http.StatusBadRequest, nil)
		return
	}
	response.OK(c, gin.H{"retried": true})
}

// Archive 归档失败任务（从 retry 迁移到 archived）。
// @Summary 归档失败任务（后台）
// @Tags admin-task
// @Produce json
// @Param queue path string true "队列名"
// @Param task_id path string true "任务ID"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/task-queues/{queue}/failed/{task_id}/archive [post]
func (h *TaskQueueHandler) Archive(c *gin.Context) {
	if h == nil || h.inspector == nil {
		handler.FailServiceUnavailable(c, nil, "asynq inspector unavailable")
		return
	}
	var uri adminreq.QueueTaskActionURI
	if err := c.ShouldBindUri(&uri); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	if err := h.inspector.ArchiveTask(strings.TrimSpace(uri.Queue), strings.TrimSpace(uri.TaskID)); err != nil {
		handler.FailByError(c, err, http.StatusBadRequest, nil)
		return
	}
	response.OK(c, gin.H{"archived": true})
}

func formatQueueTaskTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
