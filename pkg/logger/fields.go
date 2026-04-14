package logger

import (
	"go.uber.org/zap"
)

// FieldRequestID 注入全链路 RequestID。
func FieldRequestID(id string) zap.Field {
	return zap.String("request_id", id)
}
