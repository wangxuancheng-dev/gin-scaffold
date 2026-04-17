package adminhandler

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gin-scaffold/api/handler"
	adminreq "gin-scaffold/api/request/admin"
	"gin-scaffold/api/response"
	"gin-scaffold/config"
	"gin-scaffold/internal/dao"
	"gin-scaffold/internal/model"
	"gin-scaffold/middleware"
	"gin-scaffold/pkg/db"
)

const (
	fallbackAuditExportDefaultDays = 7
	fallbackAuditExportMaxDays     = 31
)

// OpsHandler 后台运维类接口。
type OpsHandler struct {
	auditDAO auditLogStore
}

// NewOpsHandler 构造后台运维 handler。
func NewOpsHandler(auditDAO auditLogStore) *OpsHandler {
	return &OpsHandler{auditDAO: auditDAO}
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

// AuditLogsExport 导出审计日志 CSV。
// @Summary 审计日志导出（后台）
// @Tags admin-ops
// @Produce text/csv
// @Param user_id query int false "用户 ID"
// @Param action query string false "动作：POST|PUT|PATCH|DELETE"
// @Param status query int false "HTTP 状态码"
// @Param path query string false "按路径模糊匹配"
// @Param request_id query string false "按请求 ID 精确匹配"
// @Param from query string false "开始时间（RFC3339）；为空时默认 now-7d"
// @Param to query string false "结束时间（RFC3339）；为空时默认 now"
// @Param limit query int false "最大导出条数（默认 5000，最大 10000）"
// @Success 200 {string} string "csv content"
// @Router /api/v1/admin/audit-logs/export [get]
func (h *OpsHandler) AuditLogsExport(c *gin.Context) {
	if h == nil || h.auditDAO == nil {
		handler.FailServiceUnavailable(c, fmt.Errorf("audit dao not configured"), "audit service unavailable")
		return
	}
	daoQuery, ok := parseAuditQuery(c, true)
	if !ok {
		return
	}
	limit := 5000
	if s := c.Query("limit"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n <= 0 || n > 10000 {
			handler.FailInvalidParam(c, fmt.Errorf("limit must be between 1 and 10000"))
			return
		}
		limit = n
	}
	rows, err := h.auditDAO.ListForExport(c.Request.Context(), daoQuery, limit)
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	h.writeExportAudit(c, daoQuery, limit, len(rows))
	filename := "audit_logs_" + time.Now().Format("20060102_150405") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("X-Export-Count", strconv.Itoa(len(rows)))
	if daoQuery.From != nil && daoQuery.To != nil {
		c.Header("X-Export-Window", daoQuery.From.Format(time.RFC3339)+"/"+daoQuery.To.Format(time.RFC3339))
	}
	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"id", "request_id", "user_id", "role", "actor_type", "action", "path", "query", "status", "latency_ms", "client_ip", "created_at"})
	for _, r := range rows {
		_ = w.Write([]string{
			strconv.FormatInt(r.ID, 10),
			r.RequestID,
			strconv.FormatInt(r.UserID, 10),
			r.Role,
			r.ActorType,
			r.Action,
			r.Path,
			r.Query,
			strconv.Itoa(r.Status),
			strconv.Itoa(r.LatencyMS),
			r.ClientIP,
			r.CreatedAt.Format(time.RFC3339),
		})
	}
	w.Flush()
}

func (h *OpsHandler) writeExportAudit(c *gin.Context, q dao.AuditLogListQuery, limit int, exported int) {
	if h == nil || h.auditDAO == nil || c == nil {
		return
	}
	row := &model.AuditLog{
		RequestID: middleware.GetRequestID(c),
		Action:    "EXPORT",
		Path:      c.Request.URL.Path,
		Status:    200,
		LatencyMS: 0,
		ClientIP:  c.ClientIP(),
		CreatedAt: time.Now(),
	}
	if cl, ok := middleware.Claims(c); ok && cl != nil {
		row.UserID = cl.UserID
		row.Role = cl.Role
		row.ActorType = "jwt"
	} else {
		row.ActorType = "anonymous"
	}
	row.Query = buildExportAuditQuery(q, limit, exported)
	_ = h.auditDAO.Create(c.Request.Context(), row)
}

func buildExportAuditQuery(q dao.AuditLogListQuery, limit int, exported int) string {
	parts := []string{
		"exported=" + strconv.Itoa(exported),
		"limit=" + strconv.Itoa(limit),
	}
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
