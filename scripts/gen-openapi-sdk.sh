#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -o "$TMP_DIR" -d ./cmd/server,./internal/api >/dev/null
cp "$TMP_DIR/swagger.yaml" ./docs/swagger.yaml
cp "$TMP_DIR/swagger.json" ./docs/swagger.json
cp "$TMP_DIR/docs.go" ./docs/docs.go

mkdir -p ./pkg/sdk/openapi
cp ./docs/swagger.json ./pkg/sdk/openapi/swagger.json
sha256sum ./docs/swagger.json | awk '{print $1}' > ./pkg/sdk/openapi/swagger.sha256
