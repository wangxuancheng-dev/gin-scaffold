#!/usr/bin/env bash
set -euo pipefail

root="${1:-.}"
target="${root}/internal/service"

if [[ ! -d "${target}" ]]; then
  echo "skip: ${target} not found"
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

"${python_cmd[@]}" - "$target" <<'PY'
import pathlib
import re
import sys

target = pathlib.Path(sys.argv[1])
pattern = re.compile(r"errors\.Is\(\s*err\s*,\s*gorm\.ErrRecordNotFound\s*\)")
allow_nil_return = re.compile(r"return\s+.*\bnil\s*$")
return_with_err = re.compile(r"return\s+.*\berr\b")

violations = []

for path in sorted(target.rglob("*.go")):
    if path.name.endswith("_test.go"):
        continue
    text = path.read_text(encoding="utf-8")
    lines = text.splitlines()
    i = 0
    while i < len(lines):
        line = lines[i]
        if not pattern.search(line):
            i += 1
            continue
        # Ignore negative guards like `!errors.Is(err, gorm.ErrRecordNotFound)`.
        if "!errors.Is" in line:
            i += 1
            continue

        # Walk the nearest block after `if errors.Is(...) { ... }`.
        depth = lines[i].count("{") - lines[i].count("}")
        block = [lines[i]]
        j = i + 1
        while j < len(lines):
            block.append(lines[j])
            depth += lines[j].count("{") - lines[j].count("}")
            if depth <= 0:
                break
            j += 1

        block_text = "\n".join(block)
        if "errcode.New(" in block_text or "notfound-ok" in block_text:
            i = j + 1
            continue

        # Backward-compat path: notfound may intentionally degrade to success.
        if any(allow_nil_return.search(line.strip()) for line in block):
            i = j + 1
            continue

        if any(return_with_err.search(line.strip()) for line in block):
            violations.append((path, i + 1, block[0].strip()))

        i = j + 1

if violations:
    print("Found gorm.ErrRecordNotFound branches without BizError mapping:")
    for path, line_no, snippet in violations:
        rel = path.as_posix()
        print(f"- {rel}:{line_no}: {snippet}")
    print()
    print("Please map not-found branches to errcode.BizError (or add `// notfound-ok` with rationale).")
    sys.exit(1)

print("ok: service notfound mapping check passed")
PY

