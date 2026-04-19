# 缓存使用（Redis + 前缀）

## `pkg/cache`

- 构造：`cache.NewFromConfig()`，前缀来自 **`platform.cache.key_prefix`**（默认会保证以 `:` 结尾）。
- 键名：用 `client.Key("segment", "sub")` 拼出完整 Redis key，避免与 Asynq、限流等键冲突。
- 读写：`GetJSON(ctx, key, dest)`、`SetJSON(ctx, key, v, ttl)`、`Del(ctx, keys...)`。
- 底层依赖全局 **`pkg/redis`** 已初始化（由 bootstrap 完成）。

## 与业务配置缓存的区别

- **系统参数**（`system_settings`）请用 **`pkg/settings`**（带短 TTL、读已发布版本），见 [平台能力](/guide/platform) 文档「系统参数」一节。
- **`pkg/cache`** 更适合：会话外缓存、排行榜、短期计算结果等通用键值。

## 代码示例（`pkg/cache`）

在 **`bootstrap.InitServer`** 已初始化全局 Redis 的前提下，业务代码中典型用法如下（键名务必经 `Key` 拼接，避免与其它子系统冲突）：

```go
import (
    "context"
    "time"

    "gin-scaffold/pkg/cache"
)

func example(ctx context.Context) error {
    c := cache.NewFromConfig()
    key := c.Key("user", "profile", "123") // => "<prefix>user:profile:123"

    type profileVO struct {
        Name string `json:"name"`
    }

    if err := c.SetJSON(ctx, key, profileVO{Name: "Ada"}, 5*time.Minute); err != nil {
        return err
    }
    var out profileVO
    if err := c.GetJSON(ctx, key, &out); err != nil {
        return err
    }
    return c.Del(ctx, key)
}
```

## 注意

- 多实例下缓存一致性问题与业务 TTL 设计需自行评估；本包不负责缓存击穿/穿透策略。
- 生产环境 Redis **建议独立 DB 或与 Asynq 分库**（配置里 `redis.db` 与 `asynq.redis_db` 可区分）。
