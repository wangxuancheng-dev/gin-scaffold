# Changelog

## [Unreleased]

### Added

- Rate limit: `middleware.LimiterKeys` with `LimiterWithBackendKeys` / `LimiterWithStoreKeys` for custom IP and route bucket keys; docs in [rate-limiting](/guide/rate-limiting).
- Rate limit: `limiter.ip_max_per_window` / `limiter.route_max_per_window` + `window_sec` for fixed-window caps in **both** `memory` (`NewStoreWithOptions`) and `redis` (`NewRedisStore`).
- PostgreSQL migrations aligned with MySQL (`migrations/postgres/schema`, `migrations/postgres/seed`).
- Integration tests for admin menu catalog and role-filtered menu tree (`data.tree`).
- Developer handbook and guide index expanded (routing, middleware, cache, queue, migration, config, env, RBAC, admin API overview, realtime, security, testing, helpers, outbound HTTP, logging).
- `http.swagger_enabled`: Swagger UI disabled by default in production.
- `metrics.allowed_networks`: optional CIDR allowlist for `/metrics` by TCP source IP.
- `scripts/integration.sh` and CI job `integration` (Docker MySQL/Redis + integration tests).
- `deploy/systemd/gin-scaffold-worker.service.example`: standalone worker unit.
- Test coverage additions for `internal/pkg/*`, `pkg/db` / `pkg/cache` / `pkg/policy` / `pkg/httpclient` / `pkg/limiter`, `internal/routes`, `internal/api/handler`, `internal/model`, `internal/console/commands`, `internal/dao` (sqlmock), `pkg/redis` (miniredis), `pkg/loginthrottle`, `internal/app/bootstrap`, `internal/service` (menu/system settings/WS/SSE/announcement/outbox dispatcher/admin create), `internal/service/authz`.
- OSS hygiene: `LICENSE` (MIT), `CONTRIBUTING.md`, `SECURITY.md`, Dependabot (`gomod` + GitHub Actions), PR template, `.editorconfig`, `scripts/quality.sh`, `make.ps1 -Target quality`; CI runs `go test -race ./...` on Linux.

### Changed

- `configs/app.prod.yaml` / `configs/app.test.yaml`: `log.channels.daily` 与 dev 对齐，供 `logger.Channel("daily", …)` 使用。
- `configs/app.prod.yaml`: limiter defaults to Redis mode; `metrics` defaults to private/loopback CIDR allowlist.
- `routes.Build` returns `error` when metrics allowlist parse fails (no panic).
- Integration scripts prebuild migrate binary and improve handler error-helper gate scanning.
- Checklist consolidated to `docs/checklist.md` as single source of truth.
- Local quality scripts added: `scripts/go-lint.sh` / `go-lint.ps1`, `scripts/go-cover.sh` / `go-cover.ps1`.
- CI `test-build` now enforces coverage gate (`COVERAGE_THRESHOLD`, default `25`) and runs the race detector on `go test -race ./...`.

### Fixed

- Security hardening: scheduler shell command execution disabled by default in production (`scheduler.shell_commands_enabled=false`), WebSocket `CheckOrigin` aligned with CORS allowlist, `Content-Disposition` filename sanitized by `strutil.AttachmentFilename`.
- Robustness: admin query binding failures now return `FailInvalidParam`; WebSocket route moved under JWT auth and user identity comes from token claims (no spoofable `uid` query param).

---
## [v0.2.0] - 2026-04-17

### Added

- System settings module:
  - New model/dao/service/handler for `system_settings`
  - Admin APIs:
    - `GET /api/v1/admin/system-settings`
    - `GET /api/v1/admin/system-settings/{id}`
    - `POST /api/v1/admin/system-settings`
    - `PUT /api/v1/admin/system-settings/{id}`
    - `DELETE /api/v1/admin/system-settings/{id}`
  - New schema migration:
    - `migrations/mysql/schema/202504171500_create_system_settings.*.sql`
  - New permission seed:
    - `sys:config:read`
    - `sys:config:write`

- Task failure governance APIs:
  - `GET /api/v1/admin/task-queues/failed`
  - `POST /api/v1/admin/task-queues/{queue}/failed/{task_id}/retry`
  - `POST /api/v1/admin/task-queues/{queue}/failed/{task_id}/archive`

- Integration test baseline:
  - New integration suite in `tests/integration`
  - Covered flows:
    - admin auth flow
    - user async export flow
    - audit async export flow
  - One-click helper script:
    - `scripts/integration.ps1`

- Codegen docs and walkthrough:
  - `docs/guide/codegen.md`
  - `docs/guide/codegen-walkthrough.md`

### Changed

- User export switched to async task only (sync endpoint removed).
- Audit export switched to async task only (sync endpoint removed).
- `cmd/gen crud` significantly enhanced:
  - field DSL:
    - `name:type`
    - `name:type[:validate]`
    - `name:type?[:validate]`
    - default value support via `default=...`
  - generation modes:
    - `--template full|simple`
  - dry-run preview:
    - `--preview-file`
    - `--preview-full`
  - alternate output directory:
    - `--out-dir` (guarded with `--no-wire` for non-root paths)
  - added regression tests:
    - `cmd/gen/main_test.go`

### Notes

- After pulling this version, run migration before startup:

```bash
go run ./cmd/migrate up --env dev
```

- For integration testing:

```powershell
.\scripts\integration.ps1 -Action all
```


