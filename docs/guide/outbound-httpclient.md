# 出站 HTTP 客户端（超时 / 重试 / 熔断）

## 作用

`pkg/httpclient` 封装对 **下游 HTTP 服务** 的调用：超时、有限次重试、熔断（`gobreaker`），避免拖垮本服务。

## 配置（`outbound:`）

与 `config.OutboundConfig` 对应，详见 [配置说明](/guide/configuration) 中 `outbound` 小节：

- `timeout_ms`、`retry_max`、`retry_backoff_ms`
- `circuit_threshold`、`circuit_open_sec`

## 初始化

在 **`bootstrap.InitServer`** 最早阶段调用 **`httpclient.InitDefault(cfg.Outbound)`**（与配置加载顺序一致）。

## 使用方式

```go
import "gin-scaffold/pkg/httpclient"

c := httpclient.Default()
if c == nil {
    // 未初始化时勿用
}
// 构造 *http.Request 后：
resp, err := c.Do(req.WithContext(ctx))
```

- **`Do`** 内部按配置重试；对可重试状态码会退避重试（实现见 `pkg/httpclient`）。
- 业务层应始终传入 **带取消的 `context.Context`**，与请求超时或上游取消联动。

## 与 `pkg/notify` 的区别

- **httpclient**：任意下游 REST/RPC HTTP。
- **notify** 的 webhook：偏通知投递语义，配置在 `platform.notify`，见 [平台能力](/guide/platform)。
