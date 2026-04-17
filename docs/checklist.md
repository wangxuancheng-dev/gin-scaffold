# Pre-Launch Checklist

每次上线前按本清单逐项确认，建议在发布单中逐条打勾。

## 1. 构建与测试

- [ ] 拉取目标发布版本代码（tag/commit 已确认）
- [ ] 执行 `go test ./...` 通过
- [ ] 执行 `go build -o bin/server ./cmd/server` 成功
- [ ] 执行 `go build -o bin/migrate ./cmd/migrate` 成功

## 2. 配置与密钥

- [ ] `configs/app.prod.yaml` 已确认（仅非敏感默认值）
- [ ] `/opt/gin-scaffold/.env.prod` 已配置敏感项
- [ ] 执行 `sh scripts/deploy/check-prod-env.sh /opt/gin-scaffold/.env.prod` 通过
- [ ] 所有 `[WARN]` 已人工确认（可接受或已修复）
- [ ] `.env.prod` 权限为 `600`

## 3. 基础设施与进程

- [ ] 数据库可连接且账号权限正确
- [ ] 线上 migration 已执行并记录版本：`./bin/migrate --env prod --driver mysql --dsn "$DB_DSN" up`
- [ ] 审计权限 seed 已执行（旧版本升级时）：`202504171420_seed_audit_permission`、`202504171430_seed_audit_export_permission`
- [ ] Redis 可连接且密码正确
- [ ] `systemd` 服务文件已更新：`/etc/systemd/system/gin-scaffold.service`
- [ ] 执行 `sudo systemctl daemon-reload`
- [ ] 执行 `sudo systemctl restart gin-scaffold`
- [ ] 执行 `sudo systemctl status gin-scaffold` 状态正常

## 4. 网关与安全

- [ ] Nginx 配置已更新：`/etc/nginx/conf.d/gin-scaffold.conf`
- [ ] 执行 `nginx -t` 通过并 `systemctl reload nginx`
- [ ] HTTPS 证书路径与域名匹配
- [ ] `/metrics` 与 `/swagger` 白名单策略已确认
- [ ] 数据库与 Redis 未暴露公网

## 5. 上线后即时巡检

- [ ] `GET /livez` 与 `GET /readyz` 返回成功
- [ ] 核心业务接口 smoke 测试通过
- [ ] 审计查询与导出权限验证通过：`audit:read` / `audit:export`
- [ ] `GET /metrics` 可按预期访问（白名单内可访问）
- [ ] 观察 10~15 分钟日志，无明显 5xx/panic/连接错误

## 6. 回滚准备

- [ ] 上一个可运行二进制已保留
- [ ] 回滚命令已预演：替换旧二进制 + `systemctl restart gin-scaffold`
- [ ] 变更记录（版本、时间、执行人）已登记

## 7. 紧急发布（Hotfix）最小清单

仅用于紧急修复，目标是最小改动、最短路径、可快速回滚。

- [ ] 变更范围已确认且仅影响必要模块
- [ ] 至少执行受影响模块测试（或最小 smoke 测试）
- [ ] 二进制构建成功：`go build -o bin/server ./cmd/server`
- [ ] 环境变量自检通过：`sh scripts/deploy/check-prod-env.sh /opt/gin-scaffold/.env.prod`
- [ ] 发布后立即验证：`/readyz` + 1~2 个核心接口
- [ ] 连续观察 10 分钟日志，无 5xx/panic 明显异常
- [ ] 回滚版本与负责人已明确，必要时可 1 分钟内回退
