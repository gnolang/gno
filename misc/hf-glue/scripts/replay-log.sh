#!/usr/bin/env bash
# Run an in-process genesis replay via `hardfork test --verbose` and capture
# the per-tx log for analysis. Exits cleanly after replay completes (no docker,
# no persistent state).
#
# Usage: ./scripts/replay-log.sh [path/to/genesis.json]
#
# Output: out/replay.log (full log) + stdout summary
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO="$(cd "$HERE/../.." && pwd)"
OUT="$HERE/out"

GENESIS="${1:-$OUT/genesis.json}"
LOG="$OUT/replay.log"

if [[ ! -f "$GENESIS" ]]; then
  echo "missing $GENESIS — run 'make fetch' first" >&2
  exit 1
fi

mkdir -p "$OUT"

echo "── genesis replay smoke-test ────────────────────────────────"
echo "  genesis: $GENESIS"
echo "  log:     $LOG"
echo ""

cd "$REPO/misc/hardfork"
go run . test \
  --genesis "$GENESIS" \
  --verbose \
  --timeout 30m 2>&1 | tee "$LOG"

echo ""
echo "── summary ──────────────────────────────────────────────────"
echo ""
printf "  OK txs:    %d\n" "$(grep -c '^  \[OK\]'   "$LOG" || true)"
printf "  FAIL txs:  %d\n" "$(grep -c '^  \[FAIL\]' "$LOG" || true)"
echo ""
echo "  Unique failure reasons:"
grep -oE 'error=[^"]+' "$LOG" 2>/dev/null \
  | sed -E 's/error=//; s/\\n.*//' \
  | sort | uniq -c | sort -rn | head -20 | sed 's/^/    /' \
  || true
echo ""
echo "  Failed packages (addpkg):"
grep '^  \[FAIL\]' "$LOG" 2>/dev/null \
  | grep -oE 'gno\.land/[pr]/[a-zA-Z0-9/_-]+' \
  | sort -u | head -30 | sed 's/^/    /' || true
echo ""
echo "Full log: $LOG"
