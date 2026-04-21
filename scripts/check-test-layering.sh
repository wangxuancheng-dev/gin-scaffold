#!/usr/bin/env bash
set -euo pipefail

root="${1:-.}"

python_cmd=()
if command -v python >/dev/null 2>&1; then
  python_cmd=(python)
elif command -v python3 >/dev/null 2>&1; then
  python_cmd=(python3)
elif command -v py >/dev/null 2>&1; then
  python_cmd=(py -3)
else
  echo "error: python/python3/py is required"
  exit 1
fi

"${python_cmd[@]}" - "$root" <<'PY'
import os
import sys

root = sys.argv[1]
errs = []

def read_start(path):
    with open(path, "r", encoding="utf-8") as f:
        return "\n".join(f.read().splitlines()[:5])

int_dir = os.path.join(root, "tests", "integration")
if os.path.isdir(int_dir):
    for dp, _, files in os.walk(int_dir):
        for fn in files:
            if not fn.endswith(".go"):
                continue
            p = os.path.join(dp, fn)
            top = read_start(p)
            rel = os.path.relpath(p, root).replace("\\", "/")
            has_tag = "//go:build integration" in top
            is_test = fn.endswith("_test.go")
            if is_test and not has_tag:
                errs.append(f"{rel}: missing //go:build integration")
            if (not is_test) and has_tag:
                errs.append(f"{rel}: non-test file should not declare integration tag")

scn_dir = os.path.join(root, "tests", "scenario")
if os.path.isdir(scn_dir):
    for dp, _, files in os.walk(scn_dir):
        for fn in files:
            if not fn.endswith("_test.go"):
                continue
            p = os.path.join(dp, fn)
            top = read_start(p)
            rel = os.path.relpath(p, root).replace("\\", "/")
            if "//go:build integration" in top:
                errs.append(f"{rel}: scenario test must not carry integration build tag")

if errs:
    print("test layering check failed:")
    for e in errs:
        print(f"- {e}")
    sys.exit(1)

print("ok: test layering check passed")
PY
