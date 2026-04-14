package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Lock 表示基于 Redis 的简单分布式锁（SET NX PX + Lua 释放）。
type Lock struct {
	key    string
	token  string
	client *redis.Client
	ttl    time.Duration
}

// TryLock 尝试获取锁，成功返回 *Lock 与 nil。
func TryLock(ctx context.Context, key string, ttl time.Duration) (*Lock, error) {
	if client == nil {
		return nil, errors.New("redis client nil")
	}
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	ok, err := client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("lock busy: %s", key)
	}
	return &Lock{key: key, token: token, client: client, ttl: ttl}, nil
}

// Unlock 使用 Lua 校验 token 后删除，防止误删他人锁。
func (l *Lock) Unlock(ctx context.Context) error {
	if l == nil {
		return nil
	}
	script := `
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
else
  return 0
end`
	res, err := l.client.Eval(ctx, script, []string{l.key}, l.token).Result()
	if err != nil {
		return err
	}
	if res == int64(0) {
		return errors.New("lock token mismatch or expired")
	}
	return nil
}

// Refresh 续约 TTL（仍校验 token）。
func (l *Lock) Refresh(ctx context.Context) error {
	if l == nil {
		return errors.New("nil lock")
	}
	script := `
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("PEXPIRE", KEYS[1], ARGV[2])
else
  return 0
end`
	ms := l.ttl.Milliseconds()
	res, err := l.client.Eval(ctx, script, []string{l.key}, l.token, ms).Result()
	if err != nil {
		return err
	}
	if res == int64(0) {
		return errors.New("lock refresh failed")
	}
	return nil
}

func randomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
