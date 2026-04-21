#!/usr/bin/env bash
set -euo pipefail

root="${1:-.}"
base="${root}/configs/app.yaml"
testf="${root}/configs/app.test.yaml"
prodf="${root}/configs/app.prod.yaml"

for f in "$base" "$testf" "$prodf"; do
  if [[ ! -f "$f" ]]; then
    echo "error: missing config file: $f"
    exit 1
  fi
done

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

"${python_cmd[@]}" - "$base" "$testf" "$prodf" <<'PY'
import sys
import yaml

def flatten_keys(obj, prefix=""):
    out = set()
    if isinstance(obj, dict):
        for k, v in obj.items():
            p = f"{prefix}.{k}" if prefix else str(k)
            out.add(p)
            out |= flatten_keys(v, p)
    return out

paths = sys.argv[1:]
data = {}
for p in paths:
    with open(p, "r", encoding="utf-8") as f:
        data[p] = yaml.safe_load(f) or {}

base = paths[0]
base_keys = flatten_keys(data[base])
ok = True

for p in paths[1:]:
    keys = flatten_keys(data[p])
    missing = sorted(base_keys - keys)
    if missing:
        ok = False
        print(f"{p}: missing keys compared to {base}:")
        for k in missing:
            print(f"  - {k}")

if not ok:
    sys.exit(1)

print("ok: config compatibility check passed")
PY
