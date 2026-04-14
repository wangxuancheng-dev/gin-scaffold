// Package response 定义统一 JSON 响应结构。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gin-scaffold/internal/pkg/errcode"
	i18nhelper "gin-scaffold/internal/pkg/i18n"
)

// Body 统一响应体。
type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// OK 成功响应。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code: errcode.OK,
		Msg:  i18nhelper.T(c, errcode.KeySuccess, "success"),
		Data: data,
	})
}

// FailHTTP 使用 HTTP 状态码与业务码返回错误。
func FailHTTP(c *gin.Context, httpStatus int, code int, msgKey, defaultMsg string) {
	c.JSON(httpStatus, Body{
		Code: code,
		Msg:  i18nhelper.T(c, msgKey, defaultMsg),
	})
}

// FailBiz 业务错误（默认 HTTP 400）。
func FailBiz(c *gin.Context, code int, msgKey, defaultMsg string) {
	FailHTTP(c, http.StatusBadRequest, code, msgKey, defaultMsg)
}
