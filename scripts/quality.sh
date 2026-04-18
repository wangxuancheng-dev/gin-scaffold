#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

echo "== gofmt =="
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
  echo "These files need gofmt:" >&2
  echo "$unformatted" >&2
  exit 1
fi

echo "== go test ./... =="
go test ./...

echo "== coverage gate =="
bash ./scripts/go-cover.sh

echo "== OK =="
