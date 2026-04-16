# 生产运行手册

> 本页来自 `docs/production-runbook.md` 的 VitePress 版本。

## 核心目标

- 可上线
- 可回滚
- 可排障

## 最小上线步骤

1. 构建并上传二进制
2. 准备 `configs/app.prod.yaml` 与 `.env.prod`
3. 安装并启用 `systemd`
4. 配置 Nginx 反向代理
5. 发布前执行数据库迁移
6. 发布后执行健康检查与 smoke test

## 迁移建议

- 先在预发验证
- 避免高峰期重型 DDL
- 失败时先停发版，不要强行继续

## 巡检项

- `systemctl status gin-scaffold`
- `/readyz`、核心接口
- 错误日志与关键指标

> 详细内容请继续维护在本页面，逐步替代旧 Markdown 文档。
