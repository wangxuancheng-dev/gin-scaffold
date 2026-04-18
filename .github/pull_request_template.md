## Summary

- 

## Test plan

- [ ] `go test ./...`
- [ ] `bash scripts/go-cover.sh`（或 `scripts/go-cover.ps1`）满足 `COVERAGE_THRESHOLD`
- [ ] 已更新 Swagger/OpenAPI 产物（若改动了 API 注释）

## Checklist

- [ ] `gofmt` 无差异
- [ ] 未引入不必要的依赖或大范围无关重构
- [ ] 涉及安全/配置变更时已对照 `docs/checklist.md`
