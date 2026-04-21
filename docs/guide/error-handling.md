# 错误码与错误响应规范

## 响应结构

统一响应体定义在 `api/response/response.go`：

- `code`: 业务错误码（成功固定 `200`）
- `msg`: 面向前端/调用方的可读错误信息（支持 i18n）
- `data`: 成功时的业务数据
- `request_id`: 请求追踪 ID
- `trace_id`: 链路追踪 ID（开启 tracing 时）

## 错误码分层约定

- `2xx/4xx/5xx`：通用错误码，语义与 HTTP 状态对齐（如 `400/401/403/404/413/429/500`）
- `1xxxx`：业务模块码段（示例：用户模块 `10001+`）

建议新模块采用固定码段，避免跨团队冲突，例如：

- 用户模块：`10000-10999`
- 订单模块：`11000-11999`
- 支付模块：`12000-12999`

## 常用错误码

- `200`：成功
- `400`：参数错误
- `401`：未授权
- `403`：禁止访问
- `404`：资源不存在
- `413`：请求体过大
- `429`：请求过于频繁
- `500`：服务内部错误

## 实践约束

- 仅在 `response` 包统一输出响应，避免散落 `c.JSON`。
- 下层抛出 `errcode.BizError`，在 handler 层映射 HTTP 状态与业务码。
- 对外 `msg` 应可读且稳定，敏感堆栈仅写日志，不直接返回给调用方。
- 关键失败日志必须带 `request_id` 与 `trace_id`，便于定位。

## Handler 错误映射约定

统一使用 `api/handler/error_helper.go` 提供的 helper，避免每个 handler 重复编写 `errors.As` 分支：

- 参数绑定/校验失败：`handler.FailInvalidParam(c, err)`
- 资源不存在（明确 404）：`handler.FailNotFound(c, "xxx not found")`
- 非业务错误（兜底）：`handler.FailInternal(c, err)`
- 业务错误映射：`handler.FailByError(c, err, <defaultStatus>, <mapping>)`

补充约定（高标准）：

- 不把“资源不存在”归类为参数错误（避免返回 `400`），统一返回 `404`
- 即使是 WebSocket/SSE 入口，在握手/协商失败前也走统一错误 helper，保持响应体一致

## Service/DAO 错误语义矩阵（推荐）

- **DAO 层**：保持技术错误原样返回（如 `gorm.ErrRecordNotFound`、驱动错误），不做 HTTP 语义判断
- **Service 层**：把可预期业务分支转成 `errcode.BizError`
  - 资源不存在 -> `errcode.NotFound`
  - 参数非法/状态不允许 -> `errcode.BadRequest` / `errcode.Conflict` / `errcode.Forbidden`
  - 基础设施异常（DB/Redis/网络）-> 原始 error 透传
- **Handler 层**：只做 HTTP 映射（`FailByError` + mapping），不再根据底层技术细节做分支

示例（`UserNotFound -> 404`，其余业务错误按默认 `400`）：

```go
handler.FailByError(c, err, http.StatusBadRequest, map[int]handler.BizMapping{
    errcode.UserNotFound: {Status: http.StatusNotFound},
})
```

### Handler 骨架（绑定 → 业务 → 成功响应）

```go
func (h *UserHandler) Get(c *gin.Context) {
    var uri struct {
        ID int64 `uri:"id" binding:"required,min=1"`
    }
    if err := c.ShouldBindUri(&uri); err != nil {
        handler.FailInvalidParam(c, err)
        return
    }
    row, err := h.svc.GetByID(c.Request.Context(), uri.ID)
    if err != nil {
        handler.FailByError(c, err, http.StatusBadRequest, map[int]handler.BizMapping{
            errcode.UserNotFound: {Status: http.StatusNotFound},
        })
        return
    }
    response.OK(c, row)
}
```

需要覆盖消息文案时，可在 `BizMapping` 中设置：

- `MsgKey`：覆盖 i18n key
- `DefaultMsg`：覆盖默认文案

## CI 自动检查

仓库包含脚本 `scripts/check-handler-error-helper.sh`，会在 CI 中扫描 `api/handler`：

- 禁止新增直接 `response.FailHTTP(...)` / `response.FailBiz(...)`（`error_helper.go` 除外）
- 检测到后直接失败，提示改用 `error_helper` 统一 helper
