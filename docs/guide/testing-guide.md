# 测试指南

## 单元测试

- 目录：`tests/unit/`。
- 运行：`go test ./tests/unit/...` 或 `go test ./...`（默认不包含 `integration` build tag）。
- 风格：表驱动 + `testify` 按需。
- 包内测试：`internal/pkg/*`、`pkg/db`、`pkg/cache`/`policy`、`pkg/httpclient`/`limiter`、`routes`、`api/handler`、`internal/console/commands` 已补充基础用例。

## 集成测试

- 目录：`tests/integration/`，文件头 `//go:build integration`。
- 运行：需环境变量与依赖服务，见 `tests/integration/README.md`。
- 一键脚本：`bash ./scripts/integration.sh`（Linux/macOS/CI）或 `scripts/integration.ps1 -Action all`（Windows）。

## CI

- `.github/workflows/ci.yml`：`test-build` 跑单测、lint、安全扫描、覆盖率门禁；`integration` 任务跑 Docker + 集成测试。

## 本地 Lint（与 CI 对齐）

- Bash：`bash ./scripts/go-lint.sh`
- PowerShell：`pwsh ./scripts/go-lint.ps1` 或 `powershell -File ./scripts/go-lint.ps1`

## 覆盖率门禁（9+ 质量目标）

- Bash：`bash ./scripts/go-cover.sh`
- PowerShell：`pwsh ./scripts/go-cover.ps1` 或 `powershell -File ./scripts/go-cover.ps1`
- 默认阈值：`25%`；可通过环境变量 `COVERAGE_THRESHOLD` 覆盖，例如 `COVERAGE_THRESHOLD=30 bash ./scripts/go-cover.sh`。
- 一键本地质量检查：`bash scripts/quality.sh`（`gofmt` + `go test ./...` + 覆盖率门禁）。
- CI `test-build` 使用同一门禁，低于阈值会失败。

## 写新集成用例建议

1. 使用 `INTEGRATION_BASE_URL` 指向已迁移、已 seed 的环境。
2. 断言统一响应信封：`code` / `msg` / `data`。
3. 租户相关接口带 `X-Tenant-ID`。

## Handler 与 Swagger

- 修改 Swagger 注释后执行 `swag init`（参数与 CI 一致），否则 CI diff 会失败。
