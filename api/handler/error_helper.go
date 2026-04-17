package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"gin-scaffold/api/response"
	"gin-scaffold/internal/pkg/errcode"
)

// BizMapping 定义指定业务码的 HTTP 映射规则。
type BizMapping struct {
	Status     int
	MsgKey     string
	DefaultMsg string
}

// FailInvalidParam 返回统一参数错误响应。
func FailInvalidParam(c *gin.Context, err error) {
	msg := "invalid parameter"
	if err != nil {
		msg = err.Error()
	}
	response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, msg)
}

// FailInternal 返回统一内部错误响应。
func FailInternal(c *gin.Context, err error) {
	msg := "internal error"
	if err != nil {
		msg = err.Error()
	}
	response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, msg)
}

// FailByError 按约定映射业务错误；非业务错误统一返回 500。
func FailByError(c *gin.Context, err error, defaultBizStatus int, mappings map[int]BizMapping) {
	var biz *errcode.BizError
	if errors.As(err, &biz) {
		if m, ok := mappings[biz.Code]; ok {
			key := biz.Key
			if m.MsgKey != "" {
				key = m.MsgKey
			}
			msg := biz.Error()
			if m.DefaultMsg != "" {
				msg = m.DefaultMsg
			}
			response.FailHTTP(c, m.Status, biz.Code, key, msg)
			return
		}
		if defaultBizStatus <= 0 {
			defaultBizStatus = http.StatusBadRequest
		}
		response.FailHTTP(c, defaultBizStatus, biz.Code, biz.Key, biz.Error())
		return
	}
	FailInternal(c, err)
}
