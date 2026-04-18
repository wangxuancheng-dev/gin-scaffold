// Package redis 封装 go-redis 客户端、通用 KV 操作与分布式锁。
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"gin-scaffold/config"
)

var client *redis.Client

// Init 初始化全局 Redis 客户端。
func Init(cfg *config.RedisConfig) error {
	if cfg == nil {
		return fmt.Errorf("redis: nil config")
	}
	client = redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	return nil
}

// Client 返回全局 Redis 客户端。
func Client() *redis.Client {
	return client
}

// Ping 健康检查。
func Ping(ctx context.Context) error {
	if client == nil {
		return fmt.Errorf("redis not initialized")
	}
	return client.Ping(ctx).Err()
}

// Get 读取字符串。
func Get(ctx context.Context, key string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("redis not initialized")
	}
	return client.Get(ctx, key).Result()
}

// Set 写入字符串与过期时间。
func Set(ctx context.Context, key string, val interface{}, ttl time.Duration) error {
	if client == nil {
		return fmt.Errorf("redis not initialized")
	}
	return client.Set(ctx, key, val, ttl).Err()
}

// SetNX 仅在键不存在时设置，成功返回 true。
func SetNX(ctx context.Context, key string, val interface{}, ttl time.Duration) (bool, error) {
	if client == nil {
		return false, fmt.Errorf("redis not initialized")
	}
	return client.SetNX(ctx, key, val, ttl).Result()
}

// Incr 将键自增 1 并返回新值。
func Incr(ctx context.Context, key string) (int64, error) {
	if client == nil {
		return 0, fmt.Errorf("redis not initialized")
	}
	return client.Incr(ctx, key).Result()
}

// Del 删除键。
func Del(ctx context.Context, keys ...string) error {
	if client == nil {
		return fmt.Errorf("redis not initialized")
	}
	return client.Del(ctx, keys...).Err()
}

// Expire 设置过期。
func Expire(ctx context.Context, key string, ttl time.Duration) error {
	if client == nil {
		return fmt.Errorf("redis not initialized")
	}
	return client.Expire(ctx, key, ttl).Err()
}

// HSet 哈希写入。
func HSet(ctx context.Context, key string, values ...interface{}) error {
	if client == nil {
		return fmt.Errorf("redis not initialized")
	}
	return client.HSet(ctx, key, values...).Err()
}

// HGet 哈希读取。
func HGet(ctx context.Context, key, field string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("redis not initialized")
	}
	return client.HGet(ctx, key, field).Result()
}
