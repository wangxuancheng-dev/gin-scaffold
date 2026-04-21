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
table_pat = re.compile(r"create\s+table\s+(?:if\s+not\s+exists\s+)?[`\"]?([a-zA-Z0-9_]+)[`\"]?", re.I)
idx_pat = re.compile(r"create\s+(?:unique\s+)?index\s+(?:if\s+not\s+exists\s+)?[`\"]?([a-zA-Z0-9_]+)[`\"]?", re.I)
drop_table_pat = re.compile(r"drop\s+table(?:\s+if\s+exists)?\s+[`\"]?([a-zA-Z0-9_]+)[`\"]?", re.I)
drop_index_pat = re.compile(r"drop\s+index(?:\s+if\s+exists)?\s+[`\"]?([a-zA-Z0-9_]+)[`\"]?", re.I)

up_artifacts = {}
down_artifacts = {}

def read(path):
    with open(path, "r", encoding="utf-8") as f:
        return f.read().lower()

def has_sql_annotation(text, key):
    return f"migration-lint:{key}" in text

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
            created_tables = set(table_pat.findall(txt))
            created_indexes = set(idx_pat.findall(txt))
            up_artifacts[m.group(1)] = {"tables": created_tables, "indexes": created_indexes}
            if "create table" in txt and "schema/" in rel:
                if "primary key" not in txt:
                    errors.append(f"{rel}: create table should define primary key")
                if " index " not in txt and " key " not in txt and " unique " not in txt and not has_sql_annotation(txt, "allow-no-index"):
                    errors.append(f"{rel}: create table without explicit index/unique key")
            if "drop table" in txt and not has_sql_annotation(txt, "allow-drop-table"):
                errors.append(f"{rel}: drop table in up migration is forbidden unless annotated with migration-lint:allow-drop-table")
            if "truncate table" in txt and not has_sql_annotation(txt, "allow-truncate"):
                errors.append(f"{rel}: truncate table in up migration is forbidden unless annotated with migration-lint:allow-truncate")
            if "begin" in txt and txt.count("alter table") >= 3 and not has_sql_annotation(txt, "allow-long-tx"):
                errors.append(f"{rel}: multiple alter table statements inside transaction may cause long lock (add migration-lint:allow-long-tx with review)")
            continue
        m = down_pat.match(rel)
        if m:
            pair[m.group(1)].add("down")
            txt = read(full)
            down_artifacts[m.group(1)] = {
                "tables": set(drop_table_pat.findall(txt)),
                "indexes": set(drop_index_pat.findall(txt)),
            }

for base, sides in sorted(pair.items()):
    if "up" not in sides or "down" not in sides:
        errors.append(f"{base}: missing {'up' if 'up' not in sides else 'down'} migration")
        continue
    created = up_artifacts.get(base, {"tables": set(), "indexes": set()})
    dropped = down_artifacts.get(base, {"tables": set(), "indexes": set()})
    missing_table_drops = sorted(t for t in created["tables"] if t not in dropped["tables"])
    if missing_table_drops:
        errors.append(f"{base}: down migration missing table drops for {', '.join(missing_table_drops)}")
    # Index rollback check is advisory when table is dropped (table drop removes indexes implicitly).
    if not dropped["tables"]:
        missing_index_drops = sorted(i for i in created["indexes"] if i not in dropped["indexes"])
        if missing_index_drops:
            errors.append(f"{base}: down migration missing index drops for {', '.join(missing_index_drops)}")

if errors:
    print("migration lint failed:")
    for e in errors:
        print(f"- {e}")
    sys.exit(1)

print("ok: migration lint check passed")
PY
