#!/usr/bin/env sh

set -eu

ENV_FILE="${1:-.env.prod}"

if [ ! -f "$ENV_FILE" ]; then
  echo "[ERROR] env file not found: $ENV_FILE"
  echo "Usage: sh scripts/deploy/check-prod-env.sh /opt/gin-scaffold/.env.prod"
  exit 1
fi

# shellcheck disable=SC1090
set -a
. "$ENV_FILE"
set +a

required_vars="APP_ENV DB_DSN REDIS_ADDR JWT_SECRET"
missing=""

for key in $required_vars; do
  eval "value=\${$key:-}"
  if [ -z "$value" ]; then
    missing="$missing $key"
  fi
done

if [ "${APP_ENV:-}" != "prod" ]; then
  echo "[WARN] APP_ENV is '${APP_ENV:-}', expected 'prod'"
fi

if [ -n "$missing" ]; then
  echo "[ERROR] missing required variables:$missing"
  exit 1
fi

warn_count=0

# Non-blocking warnings for risky values.
jwt_len=$(printf "%s" "${JWT_SECRET:-}" | wc -c | tr -d ' ')
if [ "$jwt_len" -lt 32 ]; then
  echo "[WARN] JWT_SECRET is shorter than 32 chars; use a longer random secret"
  warn_count=$((warn_count + 1))
fi

case "${JWT_SECRET:-}" in
  *please-change-me*|*change-me*|*replace-with-a-long-random-secret*)
    echo "[WARN] JWT_SECRET looks like a placeholder value"
    warn_count=$((warn_count + 1))
    ;;
esac

case "${DB_DSN:-}" in
  *127.0.0.1*|*localhost*)
    echo "[WARN] DB_DSN points to localhost; ensure this is intentional for your prod topology"
    warn_count=$((warn_count + 1))
    ;;
esac

case "${REDIS_ADDR:-}" in
  127.0.0.1*|localhost*)
    echo "[WARN] REDIS_ADDR points to localhost; ensure this is intentional for your prod topology"
    warn_count=$((warn_count + 1))
    ;;
esac

echo "[OK] env file check passed: $ENV_FILE"
echo "[OK] required vars present: $required_vars"
if [ "$warn_count" -gt 0 ]; then
  echo "[WARN] found $warn_count risk warnings (non-blocking)"
fi
