package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"
)

const nullPlaceholder = "\x00NULL\x00"

var cacheSF singleflight.Group

// ErrCacheNull 表示缓存或回源结果明确为空（穿透防护占位）。
var ErrCacheNull = errors.New("cache: null value")

// CacheGetOrSet 先读缓存，未命中时合并回源（singleflight），写入时带 TTL 抖动。
func CacheGetOrSet(ctx context.Context, key string, ttl time.Duration, dest interface{}, loader func(ctx context.Context) (interface{}, error)) error {
	if client == nil {
		return errors.New("redis client nil")
	}
	raw, err := client.Get(ctx, key).Result()
	if err == nil {
		if raw == nullPlaceholder {
			return ErrCacheNull
		}
		return json.Unmarshal([]byte(raw), dest)
	}
	if err != redis.Nil {
		return err
	}

	v, err, _ := cacheSF.Do(key, func() (interface{}, error) {
		raw2, err2 := client.Get(ctx, key).Result()
		if err2 == nil {
			if raw2 == nullPlaceholder {
				return nil, ErrCacheNull
			}
			return []byte(raw2), nil
		}
		if err2 != redis.Nil {
			return nil, err2
		}
		val, err3 := loader(ctx)
		if err3 != nil {
			return nil, err3
		}
		if val == nil {
			_ = client.Set(ctx, key, nullPlaceholder, shortNullTTL(ttl)).Err()
			return nil, ErrCacheNull
		}
		b, err4 := json.Marshal(val)
		if err4 != nil {
			return nil, err4
		}
		if err5 := client.Set(ctx, key, b, jitterTTL(ttl)).Err(); err5 != nil {
			return nil, err5
		}
		return b, nil
	})
	if err != nil {
		return err
	}
	b, ok := v.([]byte)
	if !ok || b == nil {
		return fmt.Errorf("cache: unexpected value type %T", v)
	}
	return json.Unmarshal(b, dest)
}

func jitterTTL(base time.Duration) time.Duration {
	if base <= 0 {
		return base
	}
	j := time.Duration(rand.Intn(200)) * time.Millisecond
	return base + j
}

func shortNullTTL(base time.Duration) time.Duration {
	if base <= 0 {
		return time.Minute
	}
	if base/10 < time.Second {
		return time.Minute
	}
	return base / 10
}
