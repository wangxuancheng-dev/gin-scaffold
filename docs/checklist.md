# Pre-Launch Checklist

每次上线前按本清单逐项确认，并在发布单中逐条打勾。

## 1. 构建与测试

- [ ] 拉取目标发布版本代码（tag/commit 已确认）
- [ ] 执行 `go test ./...` 通过
- [ ] 执行 `golangci-lint run` 通过（本地可用 `bash scripts/go-lint.sh` 或 `scripts/go-lint.ps1`）
- [ ] 执行覆盖率门禁（`bash scripts/go-cover.sh` 或 `scripts/go-cover.ps1`，或 `bash scripts/quality.sh`）达到阈值（默认 `25%`）
- [ ] 执行 `go build -o bin/server ./cmd/server` 成功
- [ ] 执行 `go build -o bin/migrate ./cmd/migrate` 成功

## 2. 配置与密钥

- [ ] `configs/app.prod.yaml` 已确认（仅非敏感默认值）
- [ ] `/opt/gin-scaffold/.env.prod` 已配置敏感项
- [ ] 执行 `sh scripts/deploy/check-prod-env.sh /opt/gin-scaffold/.env.prod` 通过
- [ ] 所有 `[WARN]` 已人工确认
- [ ] `.env.prod` 权限为 `600`

## 3. 基础设施与进程

- [ ] 数据库可连接且账号权限正确
- [ ] migration 已执行并记录版本
- [ ] Redis 可连接且密码正确
- [ ] API 服务与 Worker 服务均已部署并 active

## 4. 网关与安全

- [ ] `nginx -t` 通过并已 reload
- [ ] `/metrics` 访问策略符合 `metrics.allowed_networks`
- [ ] `/swagger` 在生产环境按策略关闭或仅内网/鉴权开放
- [ ] 数据库与 Redis 未暴露公网

## 5. 上线后巡检

- [ ] `GET /livez`、`GET /readyz` 正常
- [ ] 核心业务 smoke 测试通过
- [ ] 关键日志无明显 5xx/panic 异常

## 6. 代码规范核查

- [ ] 新增/修改 handler 未绕过 `api/handler/error_helper.go`
- [ ] 业务错误码与 HTTP 状态码语义一致
- [ ] 执行 `bash ./scripts/check-handler-error-helper.sh .` 通过
- [ ] 执行 `bash ./scripts/check-service-notfound-mapping.sh .` 通过

## 7. 回滚准备

- [ ] 上一可运行版本二进制已保留
- [ ] 回滚命令已预演
- [ ] 变更记录（版本/时间/执行人）已登记
