# 运维同学阅读路径

目标：可部署、可巡检、可回滚。

## 首次接手

1. [快速开始](/guide/quick-start)（了解本地可运行方式）
2. [配置说明](/guide/configuration)
3. [生产运行手册](/ops/production-runbook)
4. [上线检查清单](/checklist)

## 发布前

- 校验 `.env.prod` 必填项
- 确认迁移命令与回滚命令
- 验证 `systemd` / Nginx 配置

## 发布后

- 立即检查 `/livez`、`/readyz`
- 观察任务中心执行日志与 error 日志
- 记录版本、时间、执行人

## 运维原则

- 先可回滚，再发布
- 先影子/预发验证，再上生产
- 配置与文档保持一致，避免“口口相传”
