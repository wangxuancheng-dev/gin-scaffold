#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

threshold="${COVERAGE_THRESHOLD:-25}"
if ! [[ "$threshold" =~ ^[0-9]+([.][0-9]+)?$ ]]; then
  echo "COVERAGE_THRESHOLD must be numeric, got: $threshold" >&2
  exit 1
fi

mkdir -p coverage
profile="coverage/coverage.out"
summary="coverage/coverage.txt"

# Optional: cap parallel packages on low-RAM machines, e.g. GOTEST_PARALLEL=2 bash ./scripts/go-cover.sh
gotest_p=()
if [[ -n "${GOTEST_PARALLEL:-}" ]]; then
  gotest_p=(-p "${GOTEST_PARALLEL}")
fi

go test "${gotest_p[@]}" -covermode=atomic -coverprofile "$profile" ./...
go tool cover -func="$profile" | tee "$summary"

total=$(awk '/^total:/{print $3}' "$summary" | sed 's/%//')
if [ -z "$total" ]; then
  echo "Failed to parse total coverage" >&2
  exit 1
fi

awk -v total="$total" -v threshold="$threshold" 'BEGIN { if (total + 0 < threshold + 0) exit 1 }'
if [ $? -ne 0 ]; then
  echo "Coverage gate failed: ${total}% < ${threshold}%" >&2
  exit 1
fi

echo "Coverage gate passed: ${total}% >= ${threshold}%"