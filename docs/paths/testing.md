# 测试同学阅读路径

目标：快速建立可回归的测试基线。

## 建议阅读顺序

1. [项目简介](/guide/introduction)
2. [快速开始](/guide/quick-start)
3. [命令系统](/guide/commands)
4. [定时任务中心](/guide/scheduler)
5. [日志与可观测](/guide/logging-observability)

## 重点回归清单

- 用户管理 CRUD + 导出权限
- 任务中心 CRUD / 启停 / 手动执行 / 日志查询
- `artisan` 命令在 CLI 与任务中心一致性
- 多语言响应与错误码
- 启动配置校验 fail fast

## 常用测试命令

```bash
go test ./tests/unit/...
go test ./...
```

## 问题定位建议

- 先看 `request_id` + `trace_id`
- 再看结构化日志（app/access/error/channel）
- 最后结合 `/metrics` 与任务日志定位瓶颈
