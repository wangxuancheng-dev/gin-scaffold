#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
if ! command -v golangci-lint >/dev/null 2>&1; then
  echo "golangci-lint not found. Install: https://golangci-lint.run/welcome/install/" >&2
  exit 1
fi
exec golangci-lint run ./...