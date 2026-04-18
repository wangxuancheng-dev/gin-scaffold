package jwt

import (
	"context"
	"fmt"
	"strconv"
	"time"

	appredis "gin-scaffold/pkg/redis"
)

func refreshJTIKey(uid int64) string {
	return "jwt:refresh:jti:" + strconv.FormatInt(uid, 10)
}

// SaveRefreshJTI 保存用户当前有效 refresh jti（单设备会话语义）。
func SaveRefreshJTI(ctx context.Context, uid int64, jti string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		ttl = time.Minute
	}
	return appredis.Set(ctx, refreshJTIKey(uid), jti, ttl)
}

// ValidateRefreshJTI 校验 refresh jti 是否匹配当前有效值。
func ValidateRefreshJTI(ctx context.Context, uid int64, jti string) error {
	v, err := appredis.Get(ctx, refreshJTIKey(uid))
	if err != nil {
		return fmt.Errorf("refresh jti not found or expired")
	}
	if v != jti {
		return fmt.Errorf("refresh token rotated or revoked")
	}
	return nil
}

// ClearRefreshJTI 吊销用户当前 refresh 会话（登出时调用）。
func ClearRefreshJTI(ctx context.Context, uid int64) error {
	return appredis.Del(ctx, refreshJTIKey(uid))
}
