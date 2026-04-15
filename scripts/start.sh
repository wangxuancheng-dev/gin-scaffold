#!/usr/bin/env sh
set -e
cd "$(dirname "$0")/.."
export TZ="${TZ:-UTC}"
exec go run ./cmd/server server --env "${APP_ENV:-dev}"
