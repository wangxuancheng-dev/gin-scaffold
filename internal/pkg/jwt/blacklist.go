package jwt

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	appredis "gin-scaffold/pkg/redis"
)

func blacklistKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return "jwt:blacklist:" + hex.EncodeToString(sum[:])
}

// RevokeAccessToken 将 access token 加入黑名单，直到其过期时间。
func RevokeAccessToken(ctx context.Context, token string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		ttl = time.Minute
	}
	return appredis.Set(ctx, blacklistKey(token), "1", ttl)
}

// IsAccessTokenRevoked 检查 access token 是否已吊销。
func IsAccessTokenRevoked(ctx context.Context, token string) bool {
	v, err := appredis.Get(ctx, blacklistKey(token))
	if err != nil {
		return false
	}
	return v == "1"
}
