#!/usr/bin/env bash
set -euo pipefail

root="${1:-.}"

if [[ -d "${root}/tests/unit" ]]; then
  echo "Found deprecated tests/unit directory. Use tests/scenario instead."
  exit 1
fi

if [[ ! -d "${root}/tests/scenario" ]]; then
  echo "Found missing tests/scenario directory."
  exit 1
fi

echo "ok: test layout check passed"
