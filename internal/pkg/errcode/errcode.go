// Package errcode 定义 HTTP 与业务错误码，并支持通过 gin-contrib/i18n 翻译 msg key。
package errcode

// 标准与公共错误码（与 HTTP 状态可对应）。
const (
	OK            = 200
	BadRequest    = 400
	Unauthorized  = 401
	Forbidden     = 403
	NotFound      = 404
	InternalError = 500

	// 业务用户模块 1xxxx
	UserNotFound   = 10001
	UserExists     = 10002
	UserInvalidPwd = 10003
)

// MessageKey 对应 i18n 翻译 id。
const (
	KeySuccess      = "success"
	KeyInvalidParam = "invalid_param"
	KeyUnauthorized = "unauthorized"
	KeyInternal     = "internal_error"
	KeyUserNotFound = "user_not_found"
	KeyUserExists   = "user_exists"
)
