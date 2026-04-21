package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	govalidator "github.com/go-playground/validator/v10"

	"gin-scaffold/internal/api/response"
	"gin-scaffold/internal/pkg/errcode"
	i18nhelper "gin-scaffold/internal/pkg/i18n"
	validatorpkg "gin-scaffold/internal/pkg/validator"
)

// BizMapping 定义指定业务码的 HTTP 映射规则。
type BizMapping struct {
	Status     int
	MsgKey     string
	DefaultMsg string
}

// ValidationIssue 是单个字段的校验错误详情。
type ValidationIssue struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// FailInvalidParam 返回统一参数错误响应。
func FailInvalidParam(c *gin.Context, err error) {
	msg := "invalid parameter"
	var issues []ValidationIssue
	if err != nil {
		msg, issues = explainInvalidParam(c, err)
	}
	if len(issues) > 0 {
		response.FailHTTPWithData(c, http.StatusBadRequest, errcode.BadRequest, "", msg, gin.H{
			"errors": issues,
		})
		return
	}
	// 参数错误优先返回具体原因，避免 i18n key 覆盖掉字段级校验信息。
	response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, "", msg)
}

// FailInternal 返回统一内部错误响应。
func FailInternal(c *gin.Context, err error) {
	msg := "internal error"
	if err != nil {
		msg = err.Error()
	}
	response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, msg)
}

// FailUnauthorized 返回统一未授权响应。
func FailUnauthorized(c *gin.Context, defaultMsg string) {
	if defaultMsg == "" {
		defaultMsg = "unauthorized"
	}
	response.FailHTTP(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.KeyUnauthorized, defaultMsg)
}

// FailNotFound 返回统一 404（资源不存在）。
func FailNotFound(c *gin.Context, msg string) {
	if msg == "" {
		msg = "not found"
	}
	response.FailHTTP(c, http.StatusNotFound, errcode.NotFound, errcode.KeyNotFound, msg)
}

// FailServiceUnavailable 返回统一服务不可用响应。
func FailServiceUnavailable(c *gin.Context, err error, defaultMsg string) {
	msg := defaultMsg
	if err != nil {
		msg = err.Error()
	}
	if msg == "" {
		msg = "service unavailable"
	}
	response.FailHTTP(c, http.StatusServiceUnavailable, errcode.InternalError, errcode.KeyInternal, msg)
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

func explainInvalidParam(c *gin.Context, err error) (string, []ValidationIssue) {
	var verr govalidator.ValidationErrors
	if errors.As(err, &verr) {
		langHeader := ""
		if c != nil && c.Request != nil {
			langHeader = c.GetHeader("Accept-Language")
		}
		trans := validatorpkg.TranslatorForLang(langHeader)
		msgs := make([]string, 0, len(verr))
		issues := make([]ValidationIssue, 0, len(verr))
		for _, fieldErr := range verr {
			field := validatorpkg.NormalizeFieldName(fieldErr.Field())
			msg := validationMessageFromI18n(c, field, fieldErr.Tag(), fieldErr.Param())
			if msg == "" {
				msg = strings.TrimSpace(fieldErr.Translate(trans))
			}
			if msg == "" || strings.HasPrefix(msg, "Key: '") {
				msg = i18nhelper.TFormat(c, "validation.invalid", "", map[string]string{
					"attribute": field,
					"field":     field,
					"rule":      fieldErr.Tag(),
				})
				if strings.TrimSpace(msg) == "" {
					msg = i18nhelper.TFormat(c, "validation.invalid", defaultRuleTemplate(c, "invalid"), map[string]string{
						"attribute": field,
						"field":     field,
						"rule":      fieldErr.Tag(),
					})
				}
			}
			if msg != "" {
				msgs = append(msgs, msg)
				issues = append(issues, ValidationIssue{
					Field:   field,
					Rule:    fieldErr.Tag(),
					Message: msg,
				})
			}
		}
		if len(msgs) > 0 {
			full := i18nhelper.TFormat(c, "validation_failed", "", map[string]string{
				"details": strings.Join(msgs, "; "),
			})
			if strings.TrimSpace(full) == "" {
				prefix := "参数校验失败: "
				if strings.HasPrefix(strings.ToLower(strings.TrimSpace(langHeader)), "en") {
					prefix = "validation failed: "
				}
				full = prefix + strings.Join(msgs, "; ")
			}
			return full, issues
		}
	}
	return err.Error(), nil
}

func validationMessageFromI18n(c *gin.Context, field, rule, param string) string {
	field = strings.ToLower(strings.TrimSpace(field))
	rule = strings.ToLower(strings.TrimSpace(rule))
	if field == "" || rule == "" {
		return ""
	}
	fieldLabel := i18nhelper.T(c, "validation.attributes."+field, field)
	paramValue := strings.TrimSpace(param)
	if rule == "after_field" || rule == "same_field" {
		paramField := strings.ToLower(strings.TrimSpace(param))
		paramValue = i18nhelper.T(c, "validation.attributes."+paramField, paramField)
	}
	params := map[string]string{
		"field":     field, // backward-compatible placeholder
		"attribute": fieldLabel,
		"rule":      rule,
		"param":     paramValue,
		"min":       strings.TrimSpace(param),
		"max":       strings.TrimSpace(param),
		"len":       strings.TrimSpace(param),
		"value":     strings.TrimSpace(param),
		"other":     paramValue,
	}
	// like order: custom.field.rule > custom.*.rule > field.rule (legacy) > rule.
	if msg := i18nhelper.TFormat(c, "validation.custom."+field+"."+rule, "", params); strings.TrimSpace(msg) != "" {
		return msg
	}
	if msg := i18nhelper.TFormat(c, "validation.custom.*."+rule, "", params); strings.TrimSpace(msg) != "" {
		return msg
	}
	if msg := i18nhelper.TFormat(c, "validation."+field+"."+rule, "", params); strings.TrimSpace(msg) != "" {
		return msg
	}
	// Then rule-level template: validation.min
	if msg := i18nhelper.TFormat(c, "validation."+rule, defaultRuleTemplate(c, rule), params); strings.TrimSpace(msg) != "" {
		return msg
	}
	return ""
}

func defaultRuleTemplate(c *gin.Context, rule string) string {
	isEN := false
	if c != nil && c.Request != nil {
		lang := strings.ToLower(strings.TrimSpace(c.GetHeader("Accept-Language")))
		isEN = strings.HasPrefix(lang, "en")
	}
	switch rule {
	case "required":
		if isEN {
			return ":attribute is required"
		}
		return ":attribute 为必填项"
	case "min":
		if isEN {
			return ":attribute must be at least :min characters"
		}
		return ":attribute 长度不能少于 :min"
	case "max":
		if isEN {
			return ":attribute must be at most :max characters"
		}
		return ":attribute 长度不能超过 :max"
	case "oneof":
		if isEN {
			return ":attribute must be one of: :value"
		}
		return ":attribute 必须是以下之一: :value"
	case "email":
		if isEN {
			return ":attribute must be a valid email address"
		}
		return ":attribute 必须是有效邮箱地址"
	case "after_field":
		if isEN {
			return ":attribute must be after :other"
		}
		return ":attribute 必须晚于 :other"
	case "same_field":
		if isEN {
			return ":attribute must be the same as :other"
		}
		return ":attribute 必须与 :other 一致"
	default:
		if isEN {
			return ":attribute is invalid"
		}
		return ":attribute 参数不合法"
	}
}
