#!/usr/bin/env bash
set -euo pipefail

root="${1:-.}"
target="${root}/api/handler"

if [[ ! -d "${target}" ]]; then
  echo "skip: ${target} not found"
  exit 0
fi

if ! command -v git >/dev/null 2>&1; then
  echo "error: git command is required"
  exit 1
fi

matches="$(git -C "${root}" grep -nE 'response\.FailHTTP\(|response\.FailBiz\(' -- 'api/handler' ':!api/handler/error_helper.go' || true)"
if [[ -n "${matches}" ]]; then
  echo "Found direct response.FailHTTP/FailBiz outside error_helper:"
  echo "${matches}"
  echo
  echo "Please use api/handler/error_helper.go helpers instead."
  exit 1
fi

echo "ok: handler error helper usage check passed"
