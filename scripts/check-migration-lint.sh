#!/usr/bin/env bash
set -euo pipefail

root="${1:-.}"
dir="${root}/migrations"

if [[ ! -d "$dir" ]]; then
  echo "skip: migrations directory not found"
  exit 0
fi

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

"${python_cmd[@]}" - "$dir" <<'PY'
import os
import re
import sys
from collections import defaultdict

root = sys.argv[1]
pair = defaultdict(set)
errors = []

up_pat = re.compile(r"(.+)\.up\.sql$")
down_pat = re.compile(r"(.+)\.down\.sql$")

def read(path):
    with open(path, "r", encoding="utf-8") as f:
        return f.read().lower()

for dp, _, files in os.walk(root):
    for fn in files:
        if not fn.endswith(".sql"):
            continue
        full = os.path.join(dp, fn)
        rel = os.path.relpath(full, root).replace("\\", "/")
        m = up_pat.match(rel)
        if m:
            pair[m.group(1)].add("up")
            txt = read(full)
            if "create table" in txt and "schema/" in rel:
                if " index " not in txt and " key " not in txt and " unique " not in txt:
                    errors.append(f"{rel}: create table without explicit index/unique key")
            continue
        m = down_pat.match(rel)
        if m:
            pair[m.group(1)].add("down")

for base, sides in sorted(pair.items()):
    if "up" not in sides or "down" not in sides:
        errors.append(f"{base}: missing {'up' if 'up' not in sides else 'down'} migration")

if errors:
    print("migration lint failed:")
    for e in errors:
        print(f"- {e}")
    sys.exit(1)

print("ok: migration lint check passed")
PY
