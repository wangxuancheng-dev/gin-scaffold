# 测试指南

## 单元测试

- 目录：`tests/unit/`。
- 运行：`go test ./tests/unit/...` 或 `go test ./...`（默认 **不包含** `integration` build tag 的包）。
- 风格：表驱动 + `testify` 按需；配置校验见 `config/*_test.go`。

## 集成测试

- 目录：`tests/integration/`，文件头 **`//go:build integration`**。
- 运行：需环境变量与服务，见 `tests/integration/README.md`。
- 一键脚本：**`bash ./scripts/integration.sh`**（Linux/macOS/CI）或 **`scripts/integration.ps1 -Action all`**（Windows）。

## CI

- `.github/workflows/ci.yml`：`test-build` 跑全量单测与 lint；**`integration`** Job 跑 Docker + 集成脚本。

## 写新集成用例的建议

1. 使用 `INTEGRATION_BASE_URL` 指向已迁移、已 seed 的环境。
2. 断言统一响应信封 `code` / `msg` / `data`。
3. 租户相关接口带 `X-Tenant-ID`（与 README 中说明一致）。

## Handler 与 Swagger

- 修改 Swagger 注释后执行 `swag init`（参数与 CI 一致），否则 **CI diff 失败**。
