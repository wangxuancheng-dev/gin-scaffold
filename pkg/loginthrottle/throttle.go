package loginthrottle

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	redislib "github.com/go-redis/redis/v8"

	"gin-scaffold/config"
	appredis "gin-scaffold/pkg/redis"
)

func redisKeyPrefix(cfg config.LoginSecurityConfig, cacheKeyPrefix string) string {
	if p := strings.TrimSpace(cfg.RedisKeyPrefix); p != "" {
		if !strings.HasSuffix(p, ":") {
			p += ":"
		}
		return p
	}
	p := strings.TrimSpace(cacheKeyPrefix)
	if p != "" && !strings.HasSuffix(p, ":") {
		p += ":"
	}
	if p == "" {
		return "app:"
	}
	return p
}

func identityHash(tenant, username, clientIP string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(tenant)) + "|" + strings.ToLower(strings.TrimSpace(username)) + "|" + strings.TrimSpace(clientIP)))
	return hex.EncodeToString(sum[:])[:24]
}

func failKey(prefix, tenant, username, clientIP string) string {
	return fmt.Sprintf("%slogin:v1:fail:%s:%s", prefix, tenant, identityHash(tenant, username, clientIP))
}

func lockKey(prefix, tenant, username, clientIP string) string {
	return fmt.Sprintf("%slogin:v1:lock:%s:%s", prefix, tenant, identityHash(tenant, username, clientIP))
}

// IsLocked 判断当前是否处于锁定期（失败次数达阈值后）。
func IsLocked(ctx context.Context, sec config.LoginSecurityConfig, cachePrefix, tenant, clientIP, username string) (bool, error) {
	if !sec.Enabled {
		return false, nil
	}
	prefix := redisKeyPrefix(sec, cachePrefix)
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	_, err := appredis.Get(ctx, lockKey(prefix, tenant, username, clientIP))
	if err != nil {
		if err == redislib.Nil {
			return false, nil
		}
		return false, nil
	}
	return true, nil
}

// RegisterFailure 记录一次登录失败；返回是否因本次失败触发锁定。
func RegisterFailure(ctx context.Context, sec config.LoginSecurityConfig, cachePrefix, tenant, clientIP, username string) (nowLocked bool, err error) {
	if !sec.Enabled {
		return false, nil
	}
	prefix := redisKeyPrefix(sec, cachePrefix)
	fk := failKey(prefix, tenant, username, clientIP)
	lk := lockKey(prefix, tenant, username, clientIP)

	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	n, err := appredis.Incr(ctx, fk)
	if err != nil {
		return false, err
	}
	if n == 1 {
		ttl := time.Duration(sec.WindowSec*2) * time.Second
		if ttl < 2*time.Second {
			ttl = 2 * time.Second
		}
		_ = appredis.Expire(ctx, fk, ttl)
	}
	if int(n) >= sec.MaxFailedPerWindow {
		lockTTL := time.Duration(sec.LockoutSec) * time.Second
		if lockTTL <= 0 {
			lockTTL = time.Minute
		}
		if err := appredis.Set(ctx, lk, "1", lockTTL); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// Clear 登录成功后清理失败计数与锁定。
func Clear(ctx context.Context, sec config.LoginSecurityConfig, cachePrefix, tenant, clientIP, username string) {
	if !sec.Enabled {
		return
	}
	prefix := redisKeyPrefix(sec, cachePrefix)
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	_ = appredis.Del(ctx, failKey(prefix, tenant, username, clientIP), lockKey(prefix, tenant, username, clientIP))
}
