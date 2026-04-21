// Package response 定义统一 JSON 响应结构。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	"gin-scaffold/internal/pkg/errcode"
	i18nhelper "gin-scaffold/internal/pkg/i18n"
)

// Body 统一响应体。
type Body struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	TraceID   string      `json:"trace_id,omitempty"`
}

// OK 成功响应。
func OK(c *gin.Context, data interface{}) {
	resp := Body{
		Code: errcode.OK,
		Msg:  i18nhelper.T(c, errcode.KeySuccess, "success"),
		Data: data,
	}
	fillTrace(c, &resp)
	c.JSON(http.StatusOK, resp)
}

// FailHTTP 使用 HTTP 状态码与业务码返回错误。
func FailHTTP(c *gin.Context, httpStatus int, code int, msgKey, defaultMsg string) {
	resp := Body{
		Code: code,
		Msg:  i18nhelper.T(c, msgKey, defaultMsg),
	}
	fillTrace(c, &resp)
	c.JSON(httpStatus, resp)
}

// FailBiz 业务错误（默认 HTTP 400）。
func FailBiz(c *gin.Context, code int, msgKey, defaultMsg string) {
	FailHTTP(c, http.StatusBadRequest, code, msgKey, defaultMsg)
}

func fillTrace(c *gin.Context, resp *Body) {
	if c == nil || resp == nil {
		return
	}
	if rid := c.GetString("request_id"); rid != "" {
		resp.RequestID = rid
	}
	sc := trace.SpanFromContext(c.Request.Context()).SpanContext()
	if sc.IsValid() {
		resp.TraceID = sc.TraceID().String()
	}
}
