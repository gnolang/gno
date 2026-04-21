#!/usr/bin/env bash
# Produce a structured REPORT.md from a replay log.
#
# The goal: separate root-cause failures (sig mismatch, missing param, etc)
# from cascade failures (import errors from deps that failed earlier), so we
# can decide what to fix upstream vs ignore for testing.
#
# Usage: ./scripts/report-replay.sh [path/to/replay.log]
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="$HERE/out"
LOG="${1:-$OUT/replay.log}"
REPORT="$OUT/REPLAY-REPORT.md"

if [[ ! -f "$LOG" ]]; then
  echo "missing $LOG — run 'make replay-log' first" >&2
  exit 1
fi

mkdir -p "$OUT"

total_ok=$(grep -c '^  \[OK\]'   "$LOG" || true)
total_fail=$(grep -c '^  \[FAIL\]' "$LOG" || true)
total=$((total_ok + total_fail))

# Extract failure lines. Each [FAIL] line contains the error text inline.
# Bucket by error kind.
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
grep '^  \[FAIL\]' "$LOG" > "$tmp" || true

# --- bucketing ----------------------------------------------------------------
pubkey_mismatch=$(grep -c 'PubKey does not match' "$tmp" || true)
chain_id_err=$(grep -c 'signature verification failed' "$tmp" || true)
cascade_import=$(grep -c 'could not import' "$tmp" || true)
type_check=$(grep -c 'type check errors' "$tmp" || true)
insufficient_funds=$(grep -cE '(insufficient|out of gas|insufficient funds)' "$tmp" || true)
other=$((total_fail - pubkey_mismatch - cascade_import - type_check - insufficient_funds))
[[ $other -lt 0 ]] && other=0

# --- distinct root-cause (non-cascade) packages ------------------------------
# Cascade = "could not import gno.land/..." — the package itself is fine,
# its dep failed earlier. Surface the DEPS that are missing instead.
missing_imports=$(grep -oE 'could not import gno\.land/[^ \"\\]+' "$tmp" \
  | sort | uniq -c | sort -rn | head -20 || true)

# --- packages that failed with a non-import reason (potential root causes) ----
# Grab the addpkg package path from fail lines that do NOT mention "could not import".
root_cause_fails=$(grep -v 'could not import' "$tmp" \
  | grep -oE 'gno\.land/[pr]/[a-zA-Z0-9/_.-]+' \
  | sort -u | head -30 || true)

# --- write report ------------------------------------------------------------
{
  echo "# Genesis Replay Report"
  echo ""
  echo "_Generated $(date -u +%Y-%m-%dT%H:%M:%SZ) from ${LOG}_"
  echo ""
  echo "## Summary"
  echo ""
  echo "| Metric | Count |"
  echo "|--------|------:|"
  echo "| Total txs           | $total |"
  echo "| ✅ OK                | $total_ok |"
  echo "| ❌ Failed            | $total_fail |"
  echo ""
  echo "## Failure categories"
  echo ""
  echo "| Category | Count | Kind |"
  echo "|----------|------:|------|"
  echo "| PubKey does not match signer address       | $pubkey_mismatch | **root cause** — genesis signature mismatch |"
  echo "| Signature verification failed (chain-id)   | $chain_id_err | **root cause** — chain-id leak during sig verify |"
  echo "| Type check — \`could not import\`            | $cascade_import | **cascade** — dep package failed earlier |"
  echo "| Type check — other                          | $type_check | investigate |"
  echo "| Insufficient funds / out of gas             | $insufficient_funds | investigate |"
  echo "| Other                                       | $other | investigate |"
  echo ""
  echo "## Root-cause failures"
  echo ""
  echo "These are failures NOT caused by a missing import. If any of these are"
  echo "library packages, they cause a downstream cascade."
  echo ""
  echo '```'
  if [[ -z "$root_cause_fails" ]]; then
    echo "(none detected)"
  else
    echo "$root_cause_fails"
  fi
  echo '```'
  echo ""
  echo "## Cascade — missing imports"
  echo ""
  echo "Each line is a package that downstream txs tried to import but wasn't"
  echo "deployed. If it appears here, either (a) its deploy tx failed as a"
  echo "root cause, or (b) it's not in the genesis at all."
  echo ""
  echo '```'
  if [[ -z "$missing_imports" ]]; then
    echo "(none)"
  else
    echo "$missing_imports"
  fi
  echo '```'
  echo ""
  echo "## First 10 failure log lines (for context)"
  echo ""
  echo '```'
  head -10 "$tmp" | sed 's/\\n/\n    /g'
  echo '```'
  echo ""
  echo "## Recommendation"
  echo ""
  if [[ $pubkey_mismatch -gt 0 ]]; then
    echo "- **Fix pubkey mismatch first** ($pubkey_mismatch tx). The gnoland1 genesis"
    echo "  carries signatures whose pubkey doesn't derive to the signer address."
    echo "  Either the source genesis is malformed, or our ante-handler is"
    echo "  reading the wrong pubkey during hardfork replay."
  fi
  if [[ $chain_id_err -gt 0 ]]; then
    echo "- **Chain-id leak** ($chain_id_err tx). Genesis-mode txs are being verified"
    echo "  against the new chain id instead of the original. Check the"
    echo "  \`PastChainIDs\` handling in \`loadAppState\`."
  fi
  if [[ $cascade_import -gt $((total_fail / 2)) ]]; then
    echo "- **Most failures are cascade.** Fixing the root causes above will"
    echo "  likely drop the failure count dramatically."
  fi
  echo ""
} > "$REPORT"

echo "report written to: $REPORT"
echo ""
cat "$REPORT" | head -60
