# 架构术语词汇表

为避免团队在 `dao/service/handler`、`repository/application/transport` 之间混用，本仓库固定采用以下词汇：

- `DAO`：数据访问对象，位于 `internal/dao`，只负责持久化与查询，不做业务决策。
- `Service`：业务编排层，位于 `internal/service`，负责业务规则、事务和错误语义映射。
- `Handler`：HTTP 适配层，位于 `internal/api/handler`，负责参数绑定、调用 service、返回统一响应。
- `Request/Response`：HTTP DTO，位于 `internal/api/request` 与 `internal/api/response`。
- `Routes`：路由装配层，位于 `internal/routes`，只做路由与中间件挂载。
- `Middleware`：横切 HTTP 关注点，位于 `internal/middleware`。

## 语义边界

- `DAO` 不返回 HTTP 语义，只返回技术错误（如 `gorm.ErrRecordNotFound`）。
- `Service` 把可预期业务分支映射为 `errcode.BizError`。
- `Handler` 只做 HTTP 映射，不写 SQL 与业务规则。

## 命名规范（新增代码）

- 新模块固定命名：`<module>_dao.go`、`<module>_service.go`、`<module>_handler.go`。
- 文件和路由资源名统一 `snake_case`。
- 避免在新代码里引入 `repository/usecase/controller` 等并行术语，防止语义分裂。
