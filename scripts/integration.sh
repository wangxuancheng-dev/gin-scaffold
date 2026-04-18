#!/usr/bin/env bash
# 一键集成测试：Docker MySQL/Redis → 迁移与 fixture → 启动 server+worker → go test -tags=integration。
# 适用于 Linux/macOS 与 GitHub Actions；Windows 本地请用 scripts/integration.ps1。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

export INTEGRATION_BASE_URL="${INTEGRATION_BASE_URL:-http://127.0.0.1:18080}"
export INTEGRATION_ADMIN_USERNAME="${INTEGRATION_ADMIN_USERNAME:-admin}"
export INTEGRATION_ADMIN_PASSWORD="${INTEGRATION_ADMIN_PASSWORD:-admin123456}"
export INTEGRATION_TENANT_ID="${INTEGRATION_TENANT_ID:-}"

docker compose up -d mysql redis

cleanup() {
	set +e
	if [[ -n "${SERVER_PID:-}" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
		kill "$SERVER_PID" 2>/dev/null || true
		wait "$SERVER_PID" 2>/dev/null || true
	fi
	if [[ -n "${WORKER_PID:-}" ]] && kill -0 "$WORKER_PID" 2>/dev/null; then
		kill "$WORKER_PID" 2>/dev/null || true
		wait "$WORKER_PID" 2>/dev/null || true
	fi
	docker compose down
}
trap cleanup EXIT

echo "[integration] waiting for MySQL..."
ready=0
for _ in $(seq 1 60); do
	if docker compose exec -T mysql mysqladmin ping -h localhost -uroot -proot --silent 2>/dev/null; then
		ready=1
		break
	fi
	sleep 2
done
if [[ "$ready" != "1" ]]; then
	echo "mysql not ready"
	exit 1
fi

docker compose exec -T mysql mysql -uroot -proot -e "CREATE DATABASE IF NOT EXISTS scaffold_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
go run ./cmd/migrate up --env test --driver mysql --dsn "root:root@tcp(127.0.0.1:3306)/scaffold_test?charset=utf8mb4&parseTime=True"
go run ./cmd/migrate seed up --env test --driver mysql --dsn "root:root@tcp(127.0.0.1:3306)/scaffold_test?charset=utf8mb4&parseTime=True"
docker compose exec -T mysql mysql -uroot -proot scaffold_test <tests/integration/fixtures/base.sql

echo "[integration] starting server and worker..."
go run ./cmd/server server --env test &
SERVER_PID=$!
go run ./cmd/server worker --env test &
WORKER_PID=$!

echo "[integration] waiting for ${INTEGRATION_BASE_URL}/livez ..."
start_ts=$(date +%s)
while true; do
	if curl -sf "${INTEGRATION_BASE_URL}/livez" >/dev/null; then
		break
	fi
	now=$(date +%s)
	if ((now - start_ts > 120)); then
		echo "server not ready within 120s"
		exit 1
	fi
	sleep 2
done

go test -tags=integration ./tests/integration -v
echo "[integration] OK"
