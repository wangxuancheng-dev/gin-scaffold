# 上线检查清单

> 本页来自 `docs/checklist.md` 的 VitePress 版本。

## 构建与测试

- [ ] `go test ./...`
- [ ] `go build -o bin/server ./cmd/server`
- [ ] `go build -o bin/migrate ./cmd/migrate`

## 配置与密钥

- [ ] `configs/app.prod.yaml` 已确认
- [ ] `.env.prod` 已配置并通过检查脚本
- [ ] 敏感文件权限已收敛

## 基础设施与进程

- [ ] DB / Redis 连通
- [ ] migration 已执行并记录
- [ ] systemd 重启与状态正常

## 网关与安全

- [ ] Nginx 配置生效
- [ ] HTTPS 与白名单策略正确
- [ ] DB/Redis 未暴露公网

## 发布后巡检

- [ ] `/livez`、`/readyz` 正常
- [ ] 核心接口 smoke 通过
- [ ] 10~15 分钟无明显异常日志
