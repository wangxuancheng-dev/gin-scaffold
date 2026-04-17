package adminhandler

import (
	"context"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"

	"gin-scaffold/api/handler"
	adminreq "gin-scaffold/api/request/admin"
	"gin-scaffold/api/response"
	"gin-scaffold/config"
	"gin-scaffold/internal/dao"
	"gin-scaffold/internal/job"
	"gin-scaffold/internal/model"
	"gin-scaffold/middleware"
	"gin-scaffold/pkg/db"
	"gin-scaffold/pkg/storage"
)

const (
	fallbackAuditExportDefaultDays = 7
	fallbackAuditExportMaxDays     = 31
)

// OpsHandler 后台运维类接口。
type OpsHandler struct {
	auditDAO auditLogStore
	queue    *job.Client
}

// NewOpsHandler 构造后台运维 handler。
func NewOpsHandler(auditDAO auditLogStore, queue *job.Client) *OpsHandler {
	return &OpsHandler{auditDAO: auditDAO, queue: queue}
}

type auditLogStore interface {
	Create(ctx context.Context, row *model.AuditLog) error
	List(ctx context.Context, q dao.AuditLogListQuery) ([]model.AuditLog, int64, error)
	ListForExport(ctx context.Context, q dao.AuditLogListQuery, maxRows int) ([]model.AuditLog, error)
}

// DBPing 检查数据库连通性。
// @Summary 数据库连通性检查（后台）
// @Tags admin-ops
// @Produce json
// @Success 200 {object} response.Body
// @Router /api/v1/admin/dbping [get]
func (h *OpsHandler) DBPing(c *gin.Context) {
	if db.DB() == nil {
		handler.FailServiceUnavailable(c, nil, "db not configured")
		return
	}
	sqlDB, err := db.DB().DB()
	if err != nil {
		handler.FailServiceUnavailable(c, err, "")
		return
	}
	if err := sqlDB.PingContext(c.Request.Context()); err != nil {
		handler.FailServiceUnavailable(c, err, "")
		return
	}
	response.OK(c, gin.H{"db": "ok"})
}

// AuditLogs 分页查询审计日志。
// @Summary 审计日志查询（后台）
// @Tags admin-ops
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "分页大小（<=200）"
// @Param user_id query int false "用户 ID"
// @Param action query string false "动作：POST|PUT|PATCH|DELETE"
// @Param status query int false "HTTP 状态码"
// @Param path query string false "按路径模糊匹配"
// @Param request_id query string false "按请求 ID 精确匹配"
// @Param from query string false "开始时间（RFC3339）"
// @Param to query string false "结束时间（RFC3339）"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/audit-logs [get]
func (h *OpsHandler) AuditLogs(c *gin.Context) {
	if h == nil || h.auditDAO == nil {
		handler.FailServiceUnavailable(c, fmt.Errorf("audit dao not configured"), "audit service unavailable")
		return
	}
	daoQuery, ok := parseAuditQuery(c, false)
	if !ok {
		return
	}
	list, total, err := h.auditDAO.List(c.Request.Context(), daoQuery)
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}


// AuditLogsExportTaskCreate 创建异步导出任务（低优先级队列）。
// @Summary 创建审计日志异步导出任务（后台）
// @Tags admin-ops
// @Produce json
// @Param user_id query int false "用户 ID"
// @Param action query string false "动作：POST|PUT|PATCH|DELETE"
// @Param status query int false "HTTP 状态码"
// @Param path query string false "按路径模糊匹配"
// @Param request_id query string false "按请求 ID 精确匹配"
// @Param from query string false "开始时间（RFC3339）；为空时默认 now-export_default_days"
// @Param to query string false "结束时间（RFC3339）；为空时默认 now"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/audit-logs/export/tasks [post]
func (h *OpsHandler) AuditLogsExportTaskCreate(c *gin.Context) {
	if h == nil || h.auditDAO == nil || h.queue == nil {
		handler.FailServiceUnavailable(c, fmt.Errorf("export queue unavailable"), "export queue unavailable")
		return
	}
	q, ok := parseAuditQuery(c, true)
	if !ok {
		return
	}
	taskID := uuid.NewString()
	operator := int64(0)
	if cl, ok := middleware.Claims(c); ok && cl != nil {
		operator = cl.UserID
	}
	payload := job.AuditExportPayload{
		TaskID:    taskID,
		Operator:  operator,
		UserID:    q.UserID,
		Action:    q.Action,
		Status:    q.Status,
		PathLike:  q.PathLike,
		RequestID: q.RequestID,
	}
	if q.From != nil {
		payload.From = q.From.Format(time.RFC3339)
	}
	if q.To != nil {
		payload.To = q.To.Format(time.RFC3339)
	}
	filterSummary := buildExportFilterSummary(q)
	if err := job.SetAuditExportStatus(c.Request.Context(), &job.AuditExportStatus{
		TaskID: taskID,
		State:  "queued",
		Filter: filterSummary,
	}); err != nil {
		handler.FailInternal(c, err)
		return
	}
	if err := h.queue.EnqueueTaskInQueue(c.Request.Context(), "low", job.TypeAuditExport, payload, 0); err != nil {
		handler.FailInternal(c, err)
		return
	}
	response.OK(c, gin.H{"task_id": taskID, "state": "queued", "filter": filterSummary})
}

// AuditLogsExportTaskStatus 查询异步导出任务状态。
// @Summary 查询审计日志异步导出任务状态（后台）
// @Tags admin-ops
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/audit-logs/export/tasks/{task_id} [get]
func (h *OpsHandler) AuditLogsExportTaskStatus(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("task_id"))
	if taskID == "" {
		handler.FailInvalidParam(c, fmt.Errorf("task_id is required"))
		return
	}
	st, err := job.GetAuditExportStatus(c.Request.Context(), taskID)
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	if st == nil {
		handler.FailInvalidParam(c, fmt.Errorf("task not found"))
		return
	}
	out := buildAuditExportStatusResponse(c, st, taskID)
	response.OK(c, out)
}

// AuditLogsExportTaskDownload 下载异步导出的结果文件。
// @Summary 下载审计日志异步导出结果（后台）
// @Tags admin-ops
// @Produce text/csv
// @Param task_id path string true "任务ID"
// @Success 200 {file} file
// @Router /api/v1/admin/audit-logs/export/tasks/{task_id}/download [get]
func (h *OpsHandler) AuditLogsExportTaskDownload(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("task_id"))
	if taskID == "" {
		handler.FailInvalidParam(c, fmt.Errorf("task_id is required"))
		return
	}
	st, err := job.GetAuditExportStatus(c.Request.Context(), taskID)
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	if st == nil || st.State != "success" || strings.TrimSpace(st.FileKey) == "" {
		handler.FailInvalidParam(c, fmt.Errorf("export file is not ready"))
		return
	}
	sp, err := storage.Require()
	if err != nil {
		handler.FailServiceUnavailable(c, err, "storage not configured")
		return
	}
	rc, err := sp.Open(c.Request.Context(), st.FileKey)
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	defer rc.Close()
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=\""+path.Base(st.FileKey)+"\"")
	c.Status(200)
	_, _ = io.Copy(c.Writer, rc)
}

func buildExportFilterSummary(q dao.AuditLogListQuery) string {
	parts := make([]string, 0, 8)
	if q.UserID > 0 {
		parts = append(parts, "user_id="+strconv.FormatInt(q.UserID, 10))
	}
	if q.Action != "" {
		parts = append(parts, "action="+strings.ToUpper(strings.TrimSpace(q.Action)))
	}
	if q.Status > 0 {
		parts = append(parts, "status="+strconv.Itoa(q.Status))
	}
	if q.PathLike != "" {
		parts = append(parts, "path~="+q.PathLike)
	}
	if q.RequestID != "" {
		parts = append(parts, "request_id="+q.RequestID)
	}
	if q.From != nil {
		parts = append(parts, "from="+q.From.Format(time.RFC3339))
	}
	if q.To != nil {
		parts = append(parts, "to="+q.To.Format(time.RFC3339))
	}
	s := strings.Join(parts, "&")
	if len(s) > 1024 {
		return s[:1024]
	}
	return s
}

func parseAuditQuery(c *gin.Context, exportMode bool) (dao.AuditLogListQuery, bool) {
	var req adminreq.AuditLogListQuery
	_ = c.ShouldBindQuery(&req)
	daoQuery := dao.AuditLogListQuery{
		Page:      req.Page,
		PageSize:  req.PageSize,
		UserID:    req.UserID,
		Action:    req.Action,
		Status:    req.Status,
		PathLike:  req.Path,
		RequestID: req.RequestID,
	}
	if req.From != "" {
		t, err := time.Parse(time.RFC3339, req.From)
		if err != nil {
			handler.FailInvalidParam(c, fmt.Errorf("from must be RFC3339"))
			return dao.AuditLogListQuery{}, false
		}
		daoQuery.From = &t
	}
	if req.To != "" {
		t, err := time.Parse(time.RFC3339, req.To)
		if err != nil {
			handler.FailInvalidParam(c, fmt.Errorf("to must be RFC3339"))
			return dao.AuditLogListQuery{}, false
		}
		daoQuery.To = &t
	}
	if exportMode {
		defaultRange, maxRange := auditExportWindows()
		now := time.Now()
		if daoQuery.To == nil {
			t := now
			daoQuery.To = &t
		}
		if daoQuery.From == nil {
			t := daoQuery.To.Add(-defaultRange)
			daoQuery.From = &t
		}
		if daoQuery.From.After(*daoQuery.To) {
			handler.FailInvalidParam(c, fmt.Errorf("from must be before to"))
			return dao.AuditLogListQuery{}, false
		}
		if daoQuery.To.Sub(*daoQuery.From) > maxRange {
			handler.FailInvalidParam(c, fmt.Errorf("export time range exceeds configured max"))
			return dao.AuditLogListQuery{}, false
		}
	}
	return daoQuery, true
}

func auditExportWindows() (defaultRange time.Duration, maxRange time.Duration) {
	defaultDays := fallbackAuditExportDefaultDays
	maxDays := fallbackAuditExportMaxDays
	if cfg := config.Get(); cfg != nil {
		if cfg.Platform.Audit.ExportDefaultDays > 0 {
			defaultDays = cfg.Platform.Audit.ExportDefaultDays
		}
		if cfg.Platform.Audit.ExportMaxDays > 0 {
			maxDays = cfg.Platform.Audit.ExportMaxDays
		}
		if defaultDays > maxDays {
			defaultDays = maxDays
		}
	}
	return time.Duration(defaultDays) * 24 * time.Hour, time.Duration(maxDays) * 24 * time.Hour
}

func buildAuditExportStatusResponse(c *gin.Context, st *job.AuditExportStatus, taskID string) gin.H {
	out := gin.H{
		"task_id":       st.TaskID,
		"state":         st.State,
		"is_ready":      st.State == "success" && strings.TrimSpace(st.FileKey) != "",
		"progress_rows": st.ProgressRows,
		"file_key":      st.FileKey,
		"error":         st.Error,
		"filter":        st.Filter,
		"created_at":    st.CreatedAt,
		"updated_at":    st.UpdatedAt,
	}
	if st.State == "success" && strings.TrimSpace(st.FileKey) != "" {
		downloadPath := "/api/v1/admin/audit-logs/export/tasks/" + taskID + "/download"
		out["download_path"] = downloadPath
		out["download_url"] = buildAbsoluteURL(c, downloadPath)
	}
	return out
}

func buildAbsoluteURL(c *gin.Context, p string) string {
	if c == nil || c.Request == nil {
		return p
	}
	proto := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto"))
	if proto == "" {
		if c.Request.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	host := strings.TrimSpace(c.Request.Host)
	if host == "" {
		return p
	}
	return proto + "://" + host + p
}
