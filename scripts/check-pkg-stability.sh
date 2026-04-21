#!/usr/bin/env bash
set -euo pipefail

root="${1:-.}"
pkg_dir="${root}/pkg"
manifest="${pkg_dir}/STABILITY.yaml"

if [[ ! -d "$pkg_dir" ]]; then
  echo "skip: $pkg_dir not found"
  exit 0
fi
if [[ ! -f "$manifest" ]]; then
  echo "error: missing pkg stability manifest: $manifest"
  exit 1
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

"${python_cmd[@]}" - "$pkg_dir" "$manifest" <<'PY'
import os
import sys
import yaml

pkg_dir = sys.argv[1]
manifest = sys.argv[2]

with open(manifest, "r", encoding="utf-8") as f:
    data = yaml.safe_load(f) or {}
declared = set((data.get("packages") or {}).keys())

actual = set(
    d for d in os.listdir(pkg_dir)
    if os.path.isdir(os.path.join(pkg_dir, d))
    and d not in {"sdk", "__pycache__"}
)

missing = sorted(actual - declared)
extra = sorted(declared - actual)
if missing or extra:
    if missing:
        print("missing entries in pkg/STABILITY.yaml:")
        for m in missing:
            print(f"- {m}")
    if extra:
        print("stale entries in pkg/STABILITY.yaml:")
        for e in extra:
            print(f"- {e}")
    sys.exit(1)

allowed = {"stable", "experimental"}
for name, level in (data.get("packages") or {}).items():
    if level not in allowed:
        print(f"invalid stability level for {name}: {level}")
        sys.exit(1)

print("ok: pkg stability manifest check passed")
PY
