# Integration Tests

This directory contains opt-in integration tests for key online flows.

## Current Coverage

- Admin authentication flow:
  - login via `POST /api/v1/client/auth/login`
  - access protected admin endpoint `GET /api/v1/admin/users`
- User async export flow:
  - create task `POST /api/v1/admin/users/export/tasks`
  - poll status `GET /api/v1/admin/users/export/tasks/{task_id}`
  - download CSV `GET /api/v1/admin/users/export/tasks/{task_id}/download`
- Audit async export flow:
  - create task `POST /api/v1/admin/audit-logs/export/tasks`
  - poll status `GET /api/v1/admin/audit-logs/export/tasks/{task_id}`
  - download CSV `GET /api/v1/admin/audit-logs/export/tasks/{task_id}/download`

## Prerequisites

- Application is already running (for example in `test` env).
- Database/Redis are ready and seeded with an admin user.
- Fixture SQL is available at `tests/integration/fixtures/base.sql` for stable export data.

## Required Environment Variables

```bash
INTEGRATION_BASE_URL=http://127.0.0.1:8080
INTEGRATION_ADMIN_USERNAME=admin
INTEGRATION_ADMIN_PASSWORD=admin123456
```

## Run

```bash
go test -tags=integration ./tests/integration -v
```

If required env vars are missing, tests are skipped by design.

## One-Click Local Run (Windows / PowerShell)

```powershell
.\scripts\integration.ps1 -Action all
```

This command will:

1. start `mysql` + `redis` via Docker Compose
2. create `scaffold_test` database and run migrations (`--env test`)
3. load `tests/integration/fixtures/base.sql`
4. start API server and worker in `test` env
5. run integration tests and then stop server/worker
