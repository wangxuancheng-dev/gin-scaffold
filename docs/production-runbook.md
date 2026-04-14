# Production Runbook

本文档用于单机/小规模生产部署场景，目标是让项目可上线、可回滚、可排障。

上线动作建议同时参考：`docs/checklist.md`（含“紧急发布 Hotfix 最小清单”）。

## 1. 部署目标

- 运行方式：单机 Linux + `systemd`
- 配置来源：`configs/app.prod.yaml` + 系统环境变量（推荐 `EnvironmentFile`）
- 进程管理：`systemd` 自启动 + 自动重启
- 可观测性：应用日志 + `/livez` + `/readyz` + `/metrics`

## 2. 目录建议

```text
/opt/gin-scaffold/
  bin/server
  bin/migrate
  configs/app.yaml
  configs/app.prod.yaml
  .env.prod
```

说明：
- `configs/*.yaml` 可放模板与非敏感默认值。
- `.env.prod` 放敏感信息（数据库密码、JWT Secret 等），不要提交到 Git。

## 3. 最小上线步骤

1) 构建并上传二进制

```bash
go build -o bin/server ./cmd/server
go build -o bin/migrate ./cmd/migrate
```

2) 准备服务器配置文件

- 复制 `configs/app.prod.yaml` 到服务器并按需修改。
- 参考 `deploy/.env.prod.example` 新建 `/opt/gin-scaffold/.env.prod`，写入敏感配置。
- 用自检脚本验证必填项：

```bash
sh scripts/deploy/check-prod-env.sh /opt/gin-scaffold/.env.prod
```

说明：脚本中的 `[WARN]` 为非阻塞风险提示（如弱密钥、localhost 连接串），不影响通过。

3) 安装 `systemd` 服务文件

- 参考 `deploy/systemd/gin-scaffold.service.example`
- 放置到 `/etc/systemd/system/gin-scaffold.service`

4) 启动并设置开机自启

```bash
sudo systemctl daemon-reload
sudo systemctl enable gin-scaffold
sudo systemctl start gin-scaffold
sudo systemctl status gin-scaffold
```

5) 安装 Nginx 反向代理（推荐）

- 参考 `deploy/nginx/gin-scaffold.conf.example`
- 5 分钟操作说明：`deploy/nginx/README.md`
- 放置到 `/etc/nginx/conf.d/gin-scaffold.conf`
- 修改域名、证书路径、内网白名单网段
- 执行 `nginx -t && systemctl reload nginx`

6) 执行线上数据库迁移（发布前）

```bash
# 推荐：执行已上传的迁移二进制（线上无需 Go 环境）
./bin/migrate --env prod --driver mysql --dsn "$DB_DSN" up
```

迁移建议：

- 先在预发/影子库验证 migration 可执行与耗时
- 优先采用“先扩后缩”（expand/contract）策略，避免一次性破坏式变更
- 避免高峰期执行重型 DDL（大表加索引、改列类型等）
- 迁移失败不要强行继续发版，先回滚代码或修复 migration 后重试
- 将 migration 版本号与发布时间写入发布记录，便于审计与回滚

如需回滚上一条 migration（只回滚一步）：

```bash
./bin/migrate --env prod --driver mysql --dsn "$DB_DSN" down
```

## 4. 环境变量建议

至少设置以下变量（示例键名）：

- `APP_ENV=prod`
- `DB_DSN=...`
- `REDIS_ADDR=...`
- `REDIS_PASSWORD=...`
- `JWT_SECRET=...`

可选：

- `OTEL_EXPORTER_OTLP_ENDPOINT=...`

时区建议：

- 数据库存储统一 UTC（例如 MySQL DSN 使用 `loc=UTC`）
- API/前端展示时再转换到业务时区（如 `Asia/Shanghai`）

## 5. 健康检查与巡检

- 存活检查：`GET /livez`
- 就绪检查：`GET /readyz`
- 指标：`GET /metrics`
- Swagger：生产建议限制访问来源或仅内网开放

每次发布后至少检查：

1) `systemctl status gin-scaffold` 是否正常
2) `/readyz` 返回是否成功（依赖就绪）
3) 核心接口 smoke test 是否通过

## 6. 发布与回滚

推荐“保留上一个二进制”：

- 发布：
  - 上传新二进制到临时路径
  - 原子替换 `bin/server`
  - `systemctl restart gin-scaffold`
- 回滚：
  - 恢复上一个二进制
  - `systemctl restart gin-scaffold`

## 7. 常见故障排查

### 7.1 启动失败

- 看日志：
  - `journalctl -u gin-scaffold -n 200 --no-pager`
- 重点检查：
  - `.env.prod` 是否存在且权限正确
  - `DB_DSN`/`REDIS_ADDR` 是否可连通
  - `JWT_SECRET` 是否已设置

### 7.2 配置不生效

启动日志中会打印 `config source summary`：

- `yaml_files`：实际加载了哪些 YAML
- `dotenv_files`：实际加载了哪些 `.env` 文件（`prod` 通常为空）

如果是 `prod`，优先检查 `systemd EnvironmentFile` 与系统环境变量。

### 7.3 数据库/缓存异常

- 数据库：检查连接池与账号权限、慢 SQL 日志
- Redis：检查网络连通、密码、DB 索引配置

## 8. 安全基线（单机场景）

- `.env.prod` 文件权限为 `600`，归属运行用户
- 仅开放必要端口（通常 80/443，对外不暴露 DB/Redis）
- 反向代理（Nginx/Caddy）启用 HTTPS
- 定期备份数据库，至少保留最近 7 天

## 9. 维护建议

- 每次改配置后重启服务并验证 `/health`
- 每次发布都记录版本号、时间、执行人
- 重要参数变更先在测试环境验证
