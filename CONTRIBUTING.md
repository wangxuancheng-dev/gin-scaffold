# Contributing

感谢你愿意改进本仓库。以下为合并前建议对齐的约定（与 CI 一致）。

## 本地检查

```bash
# 等价于：gofmt + 全量测试 + 覆盖率门禁（阈值见环境变量 COVERAGE_THRESHOLD，默认见 scripts/go-cover.sh）
bash ./scripts/quality.sh
```

在 Ubuntu / macOS 上也可直接跑 CI 同款 race（需 CGO）：

```bash
go test -race ./...
```

Windows 若未启用 CGO，`-race` 可能不可用；以 GitHub Actions（Linux）结果为准。

## 提交前

1. `gofmt -w` 已格式化变更文件。
2. `go test ./...` 通过。
3. 若修改了带 `swag` 注释的 API：`swag init` 并提交 `docs/swagger*` 与 `pkg/sdk/openapi/*` 同步结果。
4. 安全相关变更请对照 `docs/checklist.md`，并在 PR 中说明影响面。

## 报告漏洞

请阅读 [SECURITY.md](./SECURITY.md)，**勿**在公开 issue 中披露可利用细节。
