package errcode

import "fmt"

// BizError 携带业务错误码与可选 i18n 参数。
type BizError struct {
	Code int
	Key  string
	Err  error
	Args map[string]interface{}
}

func (e *BizError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%d key=%s err=%v", e.Code, e.Key, e.Err)
	}
	return fmt.Sprintf("code=%d key=%s", e.Code, e.Key)
}

func (e *BizError) Unwrap() error {
	return e.Err
}

// New 构造业务错误。
func New(code int, key string) *BizError {
	return &BizError{Code: code, Key: key}
}

// Wrap 包装底层错误。
func Wrap(code int, key string, err error) *BizError {
	return &BizError{Code: code, Key: key, Err: err}
}
