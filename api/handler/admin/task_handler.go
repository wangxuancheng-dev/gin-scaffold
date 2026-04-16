package adminhandler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	adminreq "gin-scaffold/api/request/admin"
	"gin-scaffold/api/response"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/internal/pkg/validator"
	"gin-scaffold/internal/service/port"
)

type TaskHandler struct {
	svc port.ScheduledTaskService
}

func NewTaskHandler(s port.ScheduledTaskService) *TaskHandler {
	return &TaskHandler{svc: s}
}

func (h *TaskHandler) List(c *gin.Context) {
	var q adminreq.TaskListQuery
	_ = c.ShouldBindQuery(&q)
	rows, total, err := h.svc.List(c.Request.Context(), q.Page, q.PageSize)
	if err != nil {
		response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, err.Error())
		return
	}
	response.OK(c, gin.H{"total": total, "list": rows})
}

func (h *TaskHandler) Create(c *gin.Context) {
	var req adminreq.TaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	if err := validator.V().Struct(&req); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	enabled := true
	timeoutSec := 0
	concurrencyPolicy := "forbid"
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	if req.TimeoutSec != nil {
		timeoutSec = *req.TimeoutSec
	}
	if req.ConcurrencyPolicy != nil {
		concurrencyPolicy = *req.ConcurrencyPolicy
	}
	row, err := h.svc.Create(c.Request.Context(), req.Name, req.Spec, req.Command, timeoutSec, concurrencyPolicy, enabled)
	if err != nil {
		h.failWithErr(c, err)
		return
	}
	response.OK(c, row)
}

func (h *TaskHandler) Update(c *gin.Context) {
	var uri adminreq.TaskIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	var req adminreq.TaskUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	if err := validator.V().Struct(&req); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	row, err := h.svc.Update(c.Request.Context(), uri.ID, req.Name, req.Spec, req.Command, req.TimeoutSec, req.ConcurrencyPolicy, req.Enabled)
	if err != nil {
		h.failWithErr(c, err)
		return
	}
	response.OK(c, row)
}

func (h *TaskHandler) Delete(c *gin.Context) {
	var uri adminreq.TaskIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uri.ID); err != nil {
		h.failWithErr(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *TaskHandler) Toggle(c *gin.Context) {
	var uri adminreq.TaskIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	var req adminreq.TaskToggleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	if err := h.svc.SetEnabled(c.Request.Context(), uri.ID, req.Enabled); err != nil {
		h.failWithErr(c, err)
		return
	}
	response.OK(c, gin.H{"enabled": req.Enabled})
}

func (h *TaskHandler) RunNow(c *gin.Context) {
	var uri adminreq.TaskIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	if err := h.svc.RunNow(c.Request.Context(), uri.ID); err != nil {
		h.failWithErr(c, err)
		return
	}
	response.OK(c, gin.H{"run": true})
}

func (h *TaskHandler) Logs(c *gin.Context) {
	var uri adminreq.TaskIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	var q adminreq.TaskListQuery
	_ = c.ShouldBindQuery(&q)
	rows, total, err := h.svc.ListLogs(c.Request.Context(), uri.ID, q.Page, q.PageSize)
	if err != nil {
		h.failWithErr(c, err)
		return
	}
	response.OK(c, gin.H{"total": total, "list": rows})
}

func (h *TaskHandler) failWithErr(c *gin.Context, err error) {
	var biz *errcode.BizError
	if errors.As(err, &biz) {
		response.FailBiz(c, biz.Code, biz.Key, biz.Error())
		return
	}
	response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, err.Error())
}
