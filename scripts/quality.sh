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
gotest_p=()
if [[ -n "${GOTEST_PARALLEL:-}" ]]; then
  gotest_p=(-p "${GOTEST_PARALLEL}")
fi
go test "${gotest_p[@]}" ./...

echo "== handler error helper check =="
bash ./scripts/check-handler-error-helper.sh .

echo "== service notfound mapping check =="
bash ./scripts/check-service-notfound-mapping.sh .

echo "== coverage gate =="
bash ./scripts/go-cover.sh

echo "== OK =="
