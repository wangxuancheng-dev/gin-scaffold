package adminhandler

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gin-scaffold/api/handler"
	adminreq "gin-scaffold/api/request/admin"
	"gin-scaffold/api/response"
	clientresp "gin-scaffold/api/response/client"
	"gin-scaffold/internal/job"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/internal/pkg/validator"
	"gin-scaffold/internal/service/port"
	"gin-scaffold/middleware"
	"gin-scaffold/pkg/storage"
	"gin-scaffold/pkg/strutil"
)

// UserHandler 后台用户接口。
type UserHandler struct {
	svc port.UserService
	q   *job.Client
}

// NewUserHandler 构造后台用户 handler。
func NewUserHandler(s port.UserService, q ...*job.Client) *UserHandler {
	var queue *job.Client
	if len(q) > 0 {
		queue = q[0]
	}
	return &UserHandler{svc: s, q: queue}
}

// List 用户分页（后台）。
// @Summary 用户列表（后台）
// @Tags admin-user
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/users [get]
func (h *UserHandler) List(c *gin.Context) {
	var q adminreq.UserListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	rows, total, err := h.svc.List(c.Request.Context(), h.buildQuery(q), q.Page, q.PageSize)
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	list := make([]clientresp.UserVO, len(rows))
	for i := range rows {
		list[i] = clientresp.FromUser(&rows[i])
	}
	response.OK(c, gin.H{"total": total, "list": list})
}

// Get 用户详情（后台）。
// @Summary 用户详情（后台）
// @Tags admin-user
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	var uri adminreq.UserIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	u, err := h.svc.GetByID(c.Request.Context(), uri.ID)
	if err != nil {
		handler.FailByError(c, err, http.StatusNotFound, map[int]handler.BizMapping{
			errcode.UserNotFound: {Status: http.StatusNotFound},
		})
		return
	}
	response.OK(c, clientresp.FromUser(u))
}

// Create 创建用户（后台）。
// @Summary 创建用户（后台）
// @Tags admin-user
// @Accept json
// @Produce json
// @Param body body adminreq.UserCreateRequest true "创建参数"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req adminreq.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	if err := validator.V().Struct(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	u, err := h.svc.AdminCreate(c.Request.Context(), req.Username, req.Password, req.Nickname, req.Role)
	if err != nil {
		handler.FailByError(c, err, http.StatusBadRequest, nil)
		return
	}
	response.OK(c, clientresp.FromUser(u))
}

// Update 更新用户（后台）。
// @Summary 更新用户（后台）
// @Tags admin-user
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param body body adminreq.UserUpdateRequest true "更新参数"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	var uri adminreq.UserIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	var req adminreq.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	if err := validator.V().Struct(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	u, err := h.svc.AdminUpdate(c.Request.Context(), uri.ID, req.Nickname, req.Password, req.Role)
	if err != nil {
		handler.FailByError(c, err, http.StatusBadRequest, map[int]handler.BizMapping{
			errcode.UserNotFound: {Status: http.StatusNotFound},
		})
		return
	}
	response.OK(c, clientresp.FromUser(u))
}

// Delete 删除用户（后台，软删除）。
// @Summary 删除用户（后台）
// @Tags admin-user
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	var uri adminreq.UserIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	err := h.svc.AdminDelete(c.Request.Context(), uri.ID)
	if err != nil {
		handler.FailByError(c, err, http.StatusBadRequest, map[int]handler.BizMapping{
			errcode.UserNotFound: {
				Status: http.StatusNotFound,
			},
			errcode.Forbidden: {
				Status:     http.StatusForbidden,
				MsgKey:     errcode.KeySuperAdminProtected,
				DefaultMsg: "super admin cannot be deleted",
			},
		})
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

// ExportTaskCreate 创建用户异步导出任务（low 队列，CSV 全量）。
// @Summary 用户异步导出任务创建（后台）
// @Tags admin-user
// @Produce json
// @Param username query string false "用户名模糊查询"
// @Param nickname query string false "昵称模糊查询"
// @Param fields query string false "导出列，逗号分隔: id,username,nickname,created_at,role"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/users/export/tasks [post]
func (h *UserHandler) ExportTaskCreate(c *gin.Context) {
	if h == nil || h.q == nil {
		handler.FailServiceUnavailable(c, fmt.Errorf("export queue unavailable"), "export queue unavailable")
		return
	}
	var q adminreq.UserListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	fields := parseExportFields(q.Fields)
	taskID := uuid.NewString()
	operator := int64(0)
	if cl, ok := middleware.Claims(c); ok && cl != nil {
		operator = cl.UserID
	}
	payload := job.UserExportPayload{
		TaskID:   taskID,
		Operator: operator,
		Username: strings.TrimSpace(q.Username),
		Nickname: strings.TrimSpace(q.Nickname),
		Fields:   fields,
		WithRole: slices.Contains(fields, "role"),
		FileType: "csv",
	}
	filter := buildUserExportFilterSummary(payload)
	if err := job.SetUserExportStatus(c.Request.Context(), &job.UserExportStatus{
		TaskID: taskID,
		State:  "queued",
		Filter: filter,
	}); err != nil {
		handler.FailInternal(c, err)
		return
	}
	if err := h.q.EnqueueTaskInQueue(c.Request.Context(), "low", job.TypeUserExport, payload, 0); err != nil {
		handler.FailInternal(c, err)
		return
	}
	response.OK(c, gin.H{"task_id": taskID, "state": "queued", "filter": filter})
}

// ExportTaskStatus 查询用户异步导出任务状态。
// @Summary 用户异步导出任务状态（后台）
// @Tags admin-user
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} response.Body
// @Router /api/v1/admin/users/export/tasks/{task_id} [get]
func (h *UserHandler) ExportTaskStatus(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("task_id"))
	if taskID == "" {
		handler.FailInvalidParam(c, fmt.Errorf("task_id is required"))
		return
	}
	st, err := job.GetUserExportStatus(c.Request.Context(), taskID)
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	if st == nil {
		handler.FailInvalidParam(c, fmt.Errorf("task not found"))
		return
	}
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
		dp := "/api/v1/admin/users/export/tasks/" + taskID + "/download"
		out["download_path"] = dp
		out["download_url"] = buildAbsoluteURL(c, dp)
	}
	response.OK(c, out)
}

// ExportTaskDownload 下载用户异步导出结果。
// @Summary 下载用户异步导出结果（后台）
// @Tags admin-user
// @Produce text/csv
// @Param task_id path string true "任务ID"
// @Success 200 {file} file
// @Router /api/v1/admin/users/export/tasks/{task_id}/download [get]
func (h *UserHandler) ExportTaskDownload(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("task_id"))
	if taskID == "" {
		handler.FailInvalidParam(c, fmt.Errorf("task_id is required"))
		return
	}
	st, err := job.GetUserExportStatus(c.Request.Context(), taskID)
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
	c.Header("Content-Disposition", "attachment; filename=\""+strutil.AttachmentFilename(st.FileKey)+"\"")
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, rc)
}

func (h *UserHandler) buildQuery(q adminreq.UserListQuery) model.UserQuery {
	return model.UserQuery{
		Username: q.Username,
		Nickname: q.Nickname,
	}
}

func parseExportFields(s string) []string {
	allowed := map[string]struct{}{
		"id": {}, "username": {}, "nickname": {}, "created_at": {}, "role": {},
	}
	if strings.TrimSpace(s) == "" {
		return []string{"id", "username", "nickname", "created_at"}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		k := strings.ToLower(strings.TrimSpace(p))
		if _, ok := allowed[k]; !ok {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		out = append(out, k)
		seen[k] = struct{}{}
	}
	if len(out) == 0 {
		return []string{"id", "username", "nickname", "created_at"}
	}
	return out
}

func buildUserExportFilterSummary(p job.UserExportPayload) string {
	parts := []string{"file_type=csv"}
	if s := strings.TrimSpace(p.Username); s != "" {
		parts = append(parts, "username~="+s)
	}
	if s := strings.TrimSpace(p.Nickname); s != "" {
		parts = append(parts, "nickname~="+s)
	}
	if len(p.Fields) > 0 {
		parts = append(parts, "fields="+strings.Join(p.Fields, ","))
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
