// Package clientip 在请求上下文中传递客户端 IP（供 service 层防爆破等使用）。
package clientip

import "context"

type ctxKey struct{}

// With 将客户端 IP 写入 context。
func With(ctx context.Context, ip string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxKey{}, ip)
}

// FromContext 读取客户端 IP；未注入时返回空字符串。
func FromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(ctxKey{}).(string)
	return v
}
