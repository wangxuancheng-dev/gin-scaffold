# 开发同学阅读路径

目标：最快速度理解项目、完成开发、避免踩坑。

## Day 0（30-60 分钟）

1. [项目简介](/guide/introduction)
2. [快速开始](/guide/quick-start)
3. [配置说明](/guide/configuration)
4. [命令系统](/guide/commands)

## Day 1（开始编码前）

1. [定时任务中心](/guide/scheduler)
2. [日志与可观测](/guide/logging-observability)
3. `README` 中的权限矩阵与模块约定

## 开发中建议

- 本地统一使用 `dev` + `.env.local`
- 变更配置项时同步更新 `configs/*.yaml` 与文档
- 新增 artisan 命令时，优先配合任务中心复用

## 提交前自检

- `go test ./...`
- 核心接口 smoke
- 文档是否同步（命令、配置、部署项）
