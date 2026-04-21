#!/usr/bin/env bash
set -euo pipefail

root="${1:-.}"
target="${root}/pkg"

if [[ ! -d "${target}" ]]; then
  echo "skip: ${target} not found"
  exit 0
fi

if ! command -v git >/dev/null 2>&1; then
  echo "error: git command is required"
  exit 1
fi

matches="$(git -C "${root}" grep -nE '"gin-scaffold/internal/(api|routes|middleware|service|dao|app|job|model)/' -- 'pkg/**/*.go' || true)"
if [[ -n "${matches}" ]]; then
  echo "Found pkg boundary violations (pkg must not import app/domain layers):"
  echo "${matches}"
  echo
  echo "Allowed internal import scope for pkg is internal/pkg only."
  exit 1
fi

echo "ok: pkg boundary check passed"
