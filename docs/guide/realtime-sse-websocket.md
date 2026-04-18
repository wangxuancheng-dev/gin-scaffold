# SSE 与 WebSocket

当前仓库提供 **演示级** 实时通道，便于你复制扩展为业务推送（订单状态、通知等）。

## 路由（以代码为准）

| 类型 | 方法 | 路径 |
|------|------|------|
| WebSocket | GET | `/api/v1/client/ws` |
| SSE | GET | `/api/v1/client/sse/stream` |

源码：`routes/client_router.go` → `handler.WSHandler` / `handler.SSEHandler`。

## WebSocket

- 依赖 **`github.com/gorilla/websocket`**。
- 演示 Handler：`api/handler/ws_handler.go`；`CheckOrigin` 与 **`cors.allow_origins`** 对齐（`middleware.WebSocketCheckOrigin`）。`allow_origins` 为 `*` 或未配置时仍较宽松，生产请列出明确前端源或交给网关校验。
- **鉴权**：与受保护 client 路由一致，需 **`Authorization: Bearer <access>`**；连接用户 ID 来自 JWT，**不再**使用 `uid` 查询参数（避免任意冒充）。

## SSE

- `Content-Type: text/event-stream`，定时 `tick` 推送示例字符串。
- Handler：`api/handler/sse_handler.go`；业务侧可替换为订阅 `internal/service` 中 channel。

## 扩展建议

1. **鉴权**：将两路由移入 `JWTAuth` 组，或校验 query/header token。
2. **多租户**：从上下文读取 `tenant_id`，只推送该租户事件。
3. **水平扩展**：WebSocket/SSE 有状态，通常需 **sticky session** 或 **独立消息服务**（Redis Pub/Sub、NATS 等）；本演示未内置集群方案。

---

## WebSocket 生产化要点（与 Hub 实现）

### 当前 Hub 行为（`internal/pkg/websocket/hub.go`）

- **内存态**：`Hub` 保存 `userID -> Conn` 与全量连接列表；**单进程内**广播/单播。
- **单端在线策略**：同一 `uid` 再次 `Register` 会覆盖旧连接（演示用）。
- **`WSService`**（`internal/service/ws_service.go`）：封装 `BroadcastJSON` 等，供业务在任意 goroutine 调用。

### 上线前必改

| 项 | 说明 |
|----|------|
| **`Upgrader.CheckOrigin`** | 现为恒 `true`，生产必须校验 `Origin` 白名单，或仅允许内网/网关注入的域名。 |
| **鉴权** | 在 `Upgrade` 前校验 JWT / 会话；不要只靠 query 里的 `uid`。 |
| **TLS** | 对外服务走 HTTPS/WSS；反代时注意 `Connection: Upgrade` 头透传。 |

### 多副本 / 多机

- 每个 Pod 进程内 Hub **互不相通**；若需「全站广播」，常见方案：
  - **Redis Pub/Sub / Stream**：各实例订阅同一频道，收到消息后对本机 Hub 广播；
  - **消息队列**：异步任务推送通知到「通知服务」再下发 WS。
- 负载均衡器对 WS 需 **会话保持** 或 **连接级路由**，否则重连可能打到无状态节点导致重复推送或漏推（视业务而定）。

### 背压与心跳

- 为连接设置 **读/写 deadline**、**应用层 ping/pong**，避免死连接占满文件描述符。
- 广播路径上若下游处理慢，需 **有界 channel** 或丢弃策略，避免 goroutine 泄漏。

### SSE 生产注意

- 反向代理（Nginx）需关闭对 `text/event-stream` 的缓冲：`proxy_buffering off` 等。
- 长连接同样占用 FD；应设置合理超时与客户端重连策略。
