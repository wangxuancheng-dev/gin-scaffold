# Unit Tests

本目录用于放置可独立运行的单元测试示例，定位是：

- 验证核心工具与业务组件（不依赖外部 MySQL/Redis）。
- 给团队提供测试写法模板（断言、准备数据、调用、验证结果）。
- 覆盖 handler 层 `httptest + gin` 的接口校验示例。

## 运行方式

在项目根目录执行：

```bash
go test ./tests/unit/...
```

## 目前示例

- `user_response_test.go`：响应结构映射（`copier`）
- `jwt_test.go`：JWT 签发与解析
- `limiter_test.go`：限流器行为
- `user_service_login_test.go`：service 层 mock DAO（`testify/mock`）
- `user_handler_http_test.go`：handler 层 HTTP 请求测试（`httptest` + `gin`）
- `user_handler_success_test.go`：handler 层成功路径（`Register/Get/Login/List`）

或执行全量测试：

```bash
go test ./...
```
