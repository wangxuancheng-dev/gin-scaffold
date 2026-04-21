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

## 最小可复制验证

```bash
# 启动服务
go run ./cmd/server server --env dev

# 访问一个会触发缓存读写的业务接口（按你项目实际接口替换）
curl -sS "http://127.0.0.1:8080/api/v1/client/profile" \
  -H "Authorization: Bearer <client-jwt>"
```

验证点：

- 首次请求后 Redis 出现对应 `platform.cache.key_prefix` 前缀键；
- 重复请求命中缓存时，响应耗时明显下降（或日志显示命中）；
- 删除键后再次请求会回源并重建缓存。

## 常见问题与排查

- 代码写了缓存但 Redis 没有键：先确认是否使用了 `c.Key(...)` 统一拼接，以及当前环境 `redis` 连接是否正确。
- 命中率低：检查 TTL 是否过短、缓存键是否包含高频变化字段（导致键爆炸）。
- 多实例读到旧数据：写路径需配合 `Del`/更新策略，必要时缩短 TTL 或增加版本号键。
- 键冲突或难排查：统一使用业务分段键（如 `user:profile:{id}`）并保持固定前缀。
- 与队列/限流键混淆：确认 `redis.db` 与 `asynq.redis_db`、`limiter.redis_key_prefix` 的隔离策略。
