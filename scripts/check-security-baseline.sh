#!/usr/bin/env bash
set -euo pipefail

root="${1:-.}"
cfg="${root}/configs/app.prod.yaml"

if [[ ! -f "$cfg" ]]; then
  echo "error: missing production config: $cfg"
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

"${python_cmd[@]}" - "$cfg" <<'PY'
import sys
import yaml

path = sys.argv[1]
with open(path, "r", encoding="utf-8") as f:
    cfg = yaml.safe_load(f) or {}

def get(path, default=None):
    cur = cfg
    for part in path.split("."):
        if not isinstance(cur, dict) or part not in cur:
            return default
        cur = cur[part]
    return cur

violations = []
checks = [
    ("debug", False, "debug must be false in prod"),
    ("http.swagger_enabled", False, "swagger must be disabled in prod"),
    ("scheduler.shell_commands_enabled", False, "shell commands must be disabled in prod"),
    ("platform.login_security.enabled", True, "login security should be enabled in prod"),
    ("platform.idempotency.enabled", True, "idempotency should be enabled in prod"),
]

for key, want, msg in checks:
    got = get(key, None)
    if got is None:
        violations.append(f"{key}: missing ({msg})")
    elif got != want:
        violations.append(f"{key}: expected {want}, got {got} ({msg})")

origins = get("cors.allow_origins", [])
if isinstance(origins, list) and "*" in [str(x).strip() for x in origins]:
    violations.append("cors.allow_origins must not contain '*' in prod")

if violations:
    print("security baseline check failed:")
    for v in violations:
        print(f"- {v}")
    sys.exit(1)

print("ok: security baseline check passed")
PY
