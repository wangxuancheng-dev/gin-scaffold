package tenant

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

type ctxKey string

const contextKey ctxKey = "tenant_id"

func WithContext(ctx context.Context, tenantID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return ctx
	}
	return context.WithValue(ctx, contextKey, tenantID)
}

func FromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v := ctx.Value(contextKey)
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func ApplyScope(ctx context.Context, tx *gorm.DB, column string) *gorm.DB {
	if tx == nil || ctx == nil {
		return tx
	}
	tenantID := FromContext(ctx)
	if tenantID == "" {
		return tx
	}
	if strings.TrimSpace(column) == "" {
		column = "tenant_id"
	}
	return tx.Where(column+" = ?", tenantID)
}
