// Package cache 在全局 Redis 之上提供带前缀的缓存封装（JSON 编解码）。
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gin-scaffold/config"
	"gin-scaffold/pkg/redis"
)

// Client 业务缓存客户端（键 = prefix + key）。
type Client struct {
	prefix string
}

// NewFromConfig 使用当前配置中的 platform.cache.key_prefix 构造客户端。
func NewFromConfig() *Client {
	cfg := config.Get()
	prefix := "app:"
	if cfg != nil && strings.TrimSpace(cfg.Platform.Cache.KeyPrefix) != "" {
		prefix = cfg.Platform.Cache.KeyPrefix
	}
	if !strings.HasSuffix(prefix, ":") {
		prefix += ":"
	}
	return &Client{prefix: prefix}
}

// Key 拼接完整 Redis 键（使用构造时的前缀）。
func (c *Client) Key(parts ...string) string {
	p := "app:"
	if c != nil && strings.TrimSpace(c.prefix) != "" {
		p = c.prefix
	}
	return p + strings.Join(parts, ":")
}

// GetJSON 读取 JSON。
func (c *Client) GetJSON(ctx context.Context, key string, dest any) error {
	if dest == nil {
		return fmt.Errorf("cache: nil dest")
	}
	raw, err := redis.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(raw), dest)
}

// SetJSON 写入 JSON 与 TTL。
func (c *Client) SetJSON(ctx context.Context, key string, v any, ttl time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return redis.Set(ctx, key, string(b), ttl)
}

// Del 删除键。
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return redis.Del(ctx, keys...)
}
