#!/usr/bin/env bash
# audit-realm-imports.sh — scan every addpkg tx in out/source/txs.jsonl for
# imports that no longer resolve on the current gno checkout, so we catch
# realms that were valid on the source chain but now reference removed,
# renamed, or moved packages after the hardfork.
#
# Why
# ===
#   Historical addpkg txs pin imports to whatever stdlib + examples were
#   available when they were signed on mainnet. Between then and now, gno
#   has moved quickly: stdlib paths have been renamed (chain/runtime was
#   std.* before, versioned packages like gno.land/p/nt/ufmt/v0 split out,
#   some realms deleted entirely). After hardfork replay, a historical
#   realm may deserialise fine but fail the first time someone calls into
#   it because its imports no longer resolve. This audit surfaces them
#   before they surface as user-facing errors.
#
# What it reports
# ===============
#   • imports that no longer exist anywhere in the current tree
#   • per-import count across all realms importing it (bigger blast
#     radius first)
#   • a summary tally of broken vs intact realms
#
# Env
# ===
#   TXS_JSONL      default $OUT/source/txs.jsonl (historical post-genesis
#                  addpkg txs)
#   SOURCE_GENESIS default $OUT/source/config/genesis.json (base genesis-
#                  mode addpkg txs). Most realms on mainnet ship through
#                  this path, not through historical txs, so omitting it
#                  would miss the bulk of the addpkg surface.
#   REPO           gno repo root (default: two dirs up from this script)
#   OUT            misc/hf-glue/out (auto-resolved)
#
# Exit
# ====
#   0 — no broken imports found
#   1 — at least one realm imports a missing path
#   2 — prerequisite error
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HERE="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO="${REPO:-$(cd "$HERE/../.." && pwd)}"
OUT="${OUT:-$HERE/out}"
TXS_JSONL="${TXS_JSONL:-$OUT/source/txs.jsonl}"
SOURCE_GENESIS="${SOURCE_GENESIS:-$OUT/source/config/genesis.json}"

command -v jq >/dev/null 2>&1 || {
  echo "jq not found" >&2
  exit 2
}
[[ -f "$SOURCE_GENESIS" ]] || {
  echo "source genesis not found at $SOURCE_GENESIS" >&2
  exit 2
}
[[ -f "$TXS_JSONL" ]] || {
  echo "txs.jsonl not found at $TXS_JSONL" >&2
  exit 2
}

REPORT="$OUT/REALM-IMPORTS-AUDIT.md"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# ---- Build an inventory of pkgpaths that resolve on the current tree.
# Any path that exists as a directory under gnovm/stdlibs or
# examples/<realm>/ is considered resolvable.
build_known_paths() {
  # stdlib: top-level dirs under gnovm/stdlibs/ with gno files.
  # Path = the relative dir (e.g. "chain/params").
  find "$REPO/gnovm/stdlibs" -type f -name '*.gno' 2>/dev/null |
    sed -e "s|^$REPO/gnovm/stdlibs/||" -e 's|/[^/]*$||' |
    sort -u

  # examples: paths are "gno.land/..." derived from the directory tree.
  find "$REPO/examples" -type f -name '*.gno' 2>/dev/null |
    sed -e "s|^$REPO/examples/||" -e 's|/[^/]*$||' |
    sort -u
}

build_known_paths >"$WORK/known.tsv"
known_count="$(wc -l <"$WORK/known.tsv" | tr -d ' ')"

# ---- Extract every (pkgpath, imported_path) edge from addpkg txs.
# Each addpkg tx: .tx.msg[i]."@type" == "/vm.m_addpkg", msg.package.path +
# msg.package.files[*].body. Walk the bodies, pull `import "X"` lines
# (simple single-import form — does not currently handle grouped imports
# that span multiple lines; see caveat below).
#
# Source set: (a) historical txs in txs.jsonl (post-genesis addpkg calls),
# (b) genesis-mode addpkg txs in source_genesis.app_state.txs (the bulk —
# gnoland1 deployed most realms here before block 1). Both streams use the
# same amino shape so a single jq extract works on each.
jq -r '
  select(.tx.msg != null) |
  .tx.msg[] |
  select(.["@type"] == "/vm.m_addpkg") |
  .package as $p |
  $p.files[] |
  select(.name | endswith(".gno") and (endswith("_test.gno") | not) and (endswith("_filetest.gno") | not)) |
  [.name, $p.path, .body] | @tsv
' "$TXS_JSONL" >"$WORK/files.tsv" 2>/dev/null || true

jq -r '
  .app_state.txs[]? |
  select(.tx.msg != null) |
  .tx.msg[] |
  select(.["@type"] == "/vm.m_addpkg") |
  .package as $p |
  $p.files[] |
  select(.name | endswith(".gno") and (endswith("_test.gno") | not) and (endswith("_filetest.gno") | not)) |
  [.name, $p.path, .body] | @tsv
' "$SOURCE_GENESIS" >>"$WORK/files.tsv" 2>/dev/null || true

# Caveat: this parser handles two real-world styles:
#   import "path"
#   import (
#     "path1"
#     "path2"
#   )
# Single-line block imports (import ("x"; "y")) are not gno-canonical
# and are ignored.
# BSD awk (macOS default) doesn't support match()'s 3-arg array form; we
# extract quoted paths with a simple regex + substr using RSTART/RLENGTH,
# which works on both gawk and BSD awk.
awk -F'\t' '
  function extract_quoted(line,    start, len, s) {
    # returns the first "quoted" substring in line, or "" if none
    if (match(line, /"[^"]+"/)) {
      s = substr(line, RSTART + 1, RLENGTH - 2)
      return s
    }
    return ""
  }
  BEGIN { edges = 0 }
  {
    pkgpath = $2
    # File body comes in as a literal \n-separated string; split on \n.
    n = split($3, lines, /\\n/)
    in_block = 0
    for (i = 1; i <= n; i++) {
      line = lines[i]
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", line)
      if (in_block) {
        if (line == ")") { in_block = 0; continue }
        sub(/[[:space:]]*\/\/.*$/, "", line)
        q = extract_quoted(line)
        if (q != "") { print pkgpath "\t" q; edges++ }
        continue
      }
      if (line ~ /^import[[:space:]]+\(/) { in_block = 1; continue }
      if (line ~ /^import[[:space:]]+/) {
        q = extract_quoted(line)
        if (q != "") { print pkgpath "\t" q; edges++ }
      }
    }
  }
' "$WORK/files.tsv" | sort -u >"$WORK/edges.tsv"

total_edges="$(wc -l <"$WORK/edges.tsv" | tr -d ' ')"

# ---- For each imported path, decide if it resolves.
# Strip "gno.land/" prefix for examples lookup; stdlib paths have no prefix.
: >"$WORK/missing.tsv"
while IFS=$'\t' read -r pkgpath import; do
  [[ -z "$import" ]] && continue
  # Local tests sometimes import "." or relative; skip those conservatively.
  case "$import" in
  . | ./* | \.\./*) continue ;;
  esac
  # Match against known list.
  if grep -qxF "$import" "$WORK/known.tsv"; then
    continue
  fi
  printf '%s\t%s\n' "$import" "$pkgpath" >>"$WORK/missing.tsv"
done <"$WORK/edges.tsv"

missing_edges="$(wc -l <"$WORK/missing.tsv" | tr -d ' ')"

# ---- Tally: how many distinct missing imports, and which realms hit them.
cut -f1 "$WORK/missing.tsv" | sort | uniq -c | sort -rn >"$WORK/missing-by-import.tsv"
missing_imports="$(wc -l <"$WORK/missing-by-import.tsv" | tr -d ' ')"
affected_realms="$(cut -f2 "$WORK/missing.tsv" | sort -u | wc -l | tr -d ' ')"

# ---- Emit report
{
  echo "# Realm-imports audit"
  echo ""
  echo "_Generated $(date -u +%Y-%m-%dT%H:%M:%SZ)_"
  echo ""
  echo "- **Source txs**: \`$TXS_JSONL\`"
  echo "- **Tree checked**: \`$REPO/gnovm/stdlibs\` + \`$REPO/examples\`"
  echo "- **Known paths**: $known_count"
  echo "- **Edges scanned**: $total_edges (unique \`(realm, import)\` pairs)"
  echo ""
  if [[ "$missing_edges" -eq 0 ]]; then
    echo "## ✅ All historical imports resolve against the current tree"
    echo ""
    echo "Every realm deployed historically still imports packages that"
    echo "exist in this branch. No post-fork import dangling."
  else
    echo "## ❌ $missing_imports distinct missing import(s) affecting $affected_realms realm(s)"
    echo ""
    echo "Each row below is an import path that no realm-importing-it can"
    echo "resolve on this branch. If a user calls into an affected realm"
    echo "post-fork, it will fail at load time."
    echo ""
    echo "| Import path | Realms affected |"
    echo "|-------------|----------------:|"
    while read -r count path; do
      printf '| \`%s\` | %s |\n' "$path" "$count"
    done <"$WORK/missing-by-import.tsv"
    echo ""
    echo "### Realms importing missing paths"
    echo ""
    cut -f2 "$WORK/missing.tsv" | sort -u | while read -r realm; do
      printf -- '- \`%s\`\n' "$realm"
    done
  fi
} >"$REPORT"

echo "Report: $REPORT"
echo
printf 'Edges=%d Missing=%d (imports=%d, realms=%d)\n' \
  "$total_edges" "$missing_edges" "$missing_imports" "$affected_realms"

exit $((missing_edges > 0 ? 1 : 0))
