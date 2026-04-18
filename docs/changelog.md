# Changelog

## [Unreleased]

### Added

- PostgreSQL migrations aligned with MySQL (`migrations/postgres/schema`, `migrations/postgres/seed`).
- Integration tests for admin menu catalog and role-filtered menu tree (`data.tree`).

### Changed

- (placeholder) Add upcoming changes here.

### Fixed

- (placeholder) Add upcoming fixes here.

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
