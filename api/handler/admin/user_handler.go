package adminhandler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"

	"gin-scaffold/api/handler"
	adminreq "gin-scaffold/api/request/admin"
	"gin-scaffold/api/response"
	clientresp "gin-scaffold/api/response/client"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/internal/pkg/validator"
	"gin-scaffold/internal/service/port"
)

// UserHandler 后台用户接口。
type UserHandler struct {
	svc port.UserService
}

// NewUserHandler 构造后台用户 handler。
func NewUserHandler(s port.UserService) *UserHandler {
	return &UserHandler{svc: s}
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
	_ = c.ShouldBindQuery(&q)
	rows, total, err := h.svc.List(c.Request.Context(), h.buildQuery(q), q.Page, q.PageSize)
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	list := lo.Map(rows, func(u model.User, _ int) clientresp.UserVO {
		return clientresp.FromUser(&u)
	})
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

// Export 用户导出（后台）。
// @Summary 用户导出（后台）
// @Tags admin-user
// @Produce text/csv
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param username query string false "用户名（模糊）"
// @Param nickname query string false "昵称（模糊）"
// @Param export_scope query string false "导出范围: all/page, 默认 all"
// @Param export_format query string false "导出格式: csv/xlsx, 默认 csv"
// @Param export_limit query int false "全量导出上限，默认 5000"
// @Param export_batch_size query int false "流式导出批大小，默认 1000，最大 5000"
// @Param fields query string false "导出列，逗号分隔: id,username,nickname,created_at,role"
// @Success 200 {file} file "csv/xlsx file"
// @Router /api/v1/admin/users/export [get]
func (h *UserHandler) Export(c *gin.Context) {
	var q adminreq.UserListQuery
	_ = c.ShouldBindQuery(&q)

	fields := parseExportFields(q.Fields)
	pageOnly := strings.EqualFold(q.ExportScope, "page")
	withRole := slices.Contains(fields, "role")
	format := strings.ToLower(strings.TrimSpace(q.ExportFormat))
	batchSize := normalizeBatchSize(q.ExportBatchSize)
	if format == "" {
		format = "csv"
	}

	switch format {
	case "xlsx":
		if err := h.exportXLSX(c, q, fields, batchSize, pageOnly, withRole); err != nil {
			handler.FailInternal(c, err)
		}
	default:
		if err := h.exportCSV(c, q, fields, batchSize, pageOnly, withRole); err != nil {
			// Streaming started; cannot safely return JSON envelope now.
			c.Error(err)
		}
	}
}

func (h *UserHandler) exportCSV(
	c *gin.Context,
	q adminreq.UserListQuery,
	fields []string,
	batchSize int,
	pageOnly, withRole bool,
) error {
	filename := "users_" + time.Now().Format("20060102_150405") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Status(http.StatusOK)

	w := csv.NewWriter(c.Writer)
	if err := w.Write(fields); err != nil {
		return err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}

	return h.svc.StreamExport(
		c.Request.Context(),
		h.buildQuery(q),
		q.Page,
		q.PageSize,
		q.ExportLimit,
		batchSize,
		pageOnly,
		withRole,
		func(r model.UserExportRow) error {
			record := make([]string, 0, len(fields))
			for _, f := range fields {
				switch f {
				case "id":
					record = append(record, strconv.FormatInt(r.ID, 10))
				case "username":
					record = append(record, r.Username)
				case "nickname":
					record = append(record, r.Nickname)
				case "created_at":
					record = append(record, r.CreatedAt.Format(time.RFC3339))
				case "role":
					record = append(record, r.Role)
				}
			}
			if err := w.Write(record); err != nil {
				return err
			}
			w.Flush()
			if err := w.Error(); err != nil {
				return err
			}
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
			return nil
		},
	)
}

func (h *UserHandler) exportXLSX(
	c *gin.Context,
	q adminreq.UserListQuery,
	fields []string,
	batchSize int,
	pageOnly, withRole bool,
) error {
	const maxSheetRows = 1_048_576
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()

	sheetIndex := 1
	sheetName := "Sheet1"
	rowIndex := 1
	stream, err := file.NewStreamWriter(sheetName)
	if err != nil {
		return err
	}

	writeHeader := func() error {
		head := make([]interface{}, 0, len(fields))
		for _, f := range fields {
			head = append(head, f)
		}
		cell, err := excelize.CoordinatesToCellName(1, rowIndex)
		if err != nil {
			return err
		}
		if err := stream.SetRow(cell, head); err != nil {
			return err
		}
		rowIndex++
		return nil
	}
	if err := writeHeader(); err != nil {
		return err
	}

	rotateSheet := func() error {
		if err := stream.Flush(); err != nil {
			return err
		}
		sheetIndex++
		sheetName = fmt.Sprintf("Sheet%d", sheetIndex)
		file.NewSheet(sheetName)
		s, err := file.NewStreamWriter(sheetName)
		if err != nil {
			return err
		}
		stream = s
		rowIndex = 1
		return writeHeader()
	}

	err = h.svc.StreamExport(
		c.Request.Context(),
		h.buildQuery(q),
		q.Page,
		q.PageSize,
		q.ExportLimit,
		batchSize,
		pageOnly,
		withRole,
		func(r model.UserExportRow) error {
			if rowIndex > maxSheetRows {
				if err := rotateSheet(); err != nil {
					return err
				}
			}
			record := make([]interface{}, 0, len(fields))
			for _, f := range fields {
				record = append(record, exportValue(r, f))
			}
			cell, err := excelize.CoordinatesToCellName(1, rowIndex)
			if err != nil {
				return err
			}
			if err := stream.SetRow(cell, record); err != nil {
				return err
			}
			rowIndex++
			return nil
		},
	)
	if err != nil {
		return err
	}
	if err := stream.Flush(); err != nil {
		return err
	}

	filename := "users_" + time.Now().Format("20060102_150405") + ".xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Status(http.StatusOK)
	_, err = file.WriteTo(c.Writer)
	return err
}

func (h *UserHandler) buildQuery(q adminreq.UserListQuery) model.UserQuery {
	return model.UserQuery{
		Username: q.Username,
		Nickname: q.Nickname,
	}
}

func exportValue(r model.UserExportRow, field string) string {
	switch field {
	case "id":
		return strconv.FormatInt(r.ID, 10)
	case "username":
		return r.Username
	case "nickname":
		return r.Nickname
	case "created_at":
		return r.CreatedAt.Format(time.RFC3339)
	case "role":
		return r.Role
	default:
		return ""
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

func normalizeBatchSize(n int) int {
	if n <= 0 {
		return 1000
	}
	if n > 5000 {
		return 5000
	}
	return n
}
