# 测试指南

## 单元测试

- 目录：`tests/scenario/`。
- 运行：`go test ./tests/scenario/...` 或 `go test ./...`（默认不包含 `integration` build tag）。
- 风格：表驱动 + `testify` 按需。
- 包内测试：`internal/pkg/*`、`pkg/db`、`pkg/cache`/`policy`、`pkg/httpclient`/`limiter`、`internal/routes`、`internal/api/handler`、`internal/console/commands` 已补充基础用例。

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
- 规范守卫脚本：
  - `bash ./scripts/check-handler-error-helper.sh .`：禁止在 `internal/api/handler` 里绕过统一错误 helper
  - `bash ./scripts/check-service-notfound-mapping.sh .`：检查 `internal/service` 中 `gorm.ErrRecordNotFound` 分支是否映射为业务语义错误（允许带 `// notfound-ok` 的特例）
  - `bash ./scripts/check-test-layering.sh .`：强制 `tests/integration` 使用 `//go:build integration`，并禁止 `tests/scenario` 漂移成 integration
  - `bash ./scripts/check-config-compat.sh .`：检查 `app.yaml` / `app.test.yaml` / `app.prod.yaml` 的键兼容性
  - `bash ./scripts/check-security-baseline.sh .`：生产安全基线（Swagger、shell command、登录防护等）自动门禁
  - `bash ./scripts/check-pkg-stability.sh .`：校验 `pkg/STABILITY.yaml` 与真实包目录同步
  - `bash ./scripts/check-migration-lint.sh .`：迁移 up/down 对称性与基础索引风险检查

## 写新集成用例建议

1. 使用 `INTEGRATION_BASE_URL` 指向已迁移、已 seed 的环境。
2. 断言统一响应信封：`code` / `msg` / `data`。
3. 租户相关接口带 `X-Tenant-ID`。

## Handler 与 Swagger

- 修改 Swagger 注释后执行 `swag init`（参数与 CI 一致），否则 CI diff 会失败。

## 单元测试片段（`testify`）

```go
import (
    "testing"

    "github.com/stretchr/testify/require"
)

func TestFoo(t *testing.T) {
    cases := []struct {
        name string
        in   int
        want int
    }{
        {"positive", 2, 4},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            require.Equal(t, tc.want, tc.in*2)
        })
    }
}
```

## 集成测试响应断言（思路）

对 `INTEGRATION_BASE_URL` 发 HTTP 后，解析 JSON 根对象，断言 **`code`**、**`data`**（见 `internal/api/response.Body`）。租户相关接口在 Header 增加 **`X-Tenant-ID`**。完整环境变量表见 `tests/integration/README.md`。
