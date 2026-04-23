#!/usr/bin/env bash
# verify-txs-jsonl.sh — assert our cached historical tx export in
# $OUT/source/txs.jsonl matches the source chain's own book-keeping through
# HALT_HEIGHT.
#
# Why: tx-archive has no integrity check of its own. A partial fetch, a
# flaky RPC, or a bug in a future tx-archive version can silently drop txs
# and produce an export that looks fine in `wc -l` but diverges from
# reality. The hardforked genesis replays from this export, so any missing
# tx means missing state post-fork.
#
# Checks
# ======
#   1. Cardinality: block `total_txs` at HALT_HEIGHT (cumulative, from the
#      block header) must equal the number of non-blank lines in txs.jsonl.
#      Mismatch here is a hard fail.
#   2. Spot-check ($SPOT_COUNT random heights in 1..HALT_HEIGHT): for each,
#      compare the source chain's `num_txs` to the count of jsonl entries
#      with matching metadata.BlockHeight. Any mismatch is a hard fail.
#
# Env
# ===
#   RPC_URL      source-chain RPC (default https://rpc.gno.land)
#   HALT_HEIGHT  cutoff height (required; last height included in replay)
#   TXS_JSONL    path to txs.jsonl (default $OUT/source/txs.jsonl; OUT auto-
#                resolved from this script's location when not provided)
#   SPOT_COUNT   how many random heights to sample (default 10)
#   SPOT_SEED    RNG seed for spot-check reproducibility (default: $RANDOM)
#
# Exit status
# ===========
#   0  — everything matches
#   1  — divergence detected (details printed)
#   2  — prerequisite error (missing file, unreachable RPC, missing jq)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="${OUT:-$(cd "$SCRIPT_DIR/.." && pwd)/out}"

RPC_URL="${RPC_URL:-https://rpc.gno.land}"
TXS_JSONL="${TXS_JSONL:-$OUT/source/txs.jsonl}"
SPOT_COUNT="${SPOT_COUNT:-10}"
SPOT_SEED="${SPOT_SEED:-$RANDOM}"

: "${HALT_HEIGHT:?HALT_HEIGHT is required}"

command -v jq >/dev/null 2>&1 || {
  echo "jq not found on PATH" >&2
  exit 2
}
command -v curl >/dev/null 2>&1 || {
  echo "curl not found on PATH" >&2
  exit 2
}
[[ -f "$TXS_JSONL" ]] || {
  echo "txs.jsonl not found at $TXS_JSONL" >&2
  exit 2
}

echo "━━━ verify-txs-jsonl ━━━"
echo "  rpc          $RPC_URL"
echo "  halt_height  $HALT_HEIGHT"
echo "  txs.jsonl    $TXS_JSONL"
echo "  spot count   $SPOT_COUNT"
echo "  spot seed    $SPOT_SEED"
echo

fail=0

# ---- Helper: fetch block header field at a given height
rpc_block_header_field() {
  local height="$1" field="$2"
  curl -sf --max-time 30 "${RPC_URL%/}/block?height=$height" |
    jq -r ".result.block.header.$field // empty"
}

# ---- Check 1: cardinality
local_count="$(grep -cve '^[[:space:]]*$' "$TXS_JSONL")"
rpc_total="$(rpc_block_header_field "$HALT_HEIGHT" total_txs)"

if [[ -z "$rpc_total" ]]; then
  echo "  [FAIL] RPC returned no total_txs at height=$HALT_HEIGHT" >&2
  exit 1
fi

printf '  cardinality: rpc=%s  local=%s  ' "$rpc_total" "$local_count"
if [[ "$rpc_total" == "$local_count" ]]; then
  echo '[OK]'
else
  diff=$((rpc_total - local_count))
  printf '[FAIL] diff=%+d (rpc - local)\n' "$diff"
  fail=1
fi

# ---- Check 2: spot-check random heights with txs
# Pick heights from the set of BlockHeights that actually have at least one
# tx in the jsonl. This is where divergence can hide — empty-block heights
# are trivially 0==0 on both sides and don't test anything. We also include
# one height with no txs as a sanity anchor (ensures the script can observe
# zero correctly).
#
# srand with explicit seed → deterministic sampling for reproducibility.
heights_csv="$(awk -v seed="$SPOT_SEED" -v n="$SPOT_COUNT" -v hi="$HALT_HEIGHT" '
  BEGIN { srand(seed) }
  /"block_height":/ {
    match($0, /"block_height":"[0-9]+"/)
    if (RSTART) {
      v = substr($0, RSTART, RLENGTH)
      gsub(/[^0-9]/, "", v)
      h[v]++
    }
  }
  END {
    # Collect unique heights with txs, shuffle, pick first n-1
    i = 0
    for (v in h) keys[i++] = v
    # Fisher-Yates
    for (j = i - 1; j > 0; j--) {
      k = int(rand() * (j + 1))
      tmp = keys[j]; keys[j] = keys[k]; keys[k] = tmp
    }
    need = (n - 1 < i) ? n - 1 : i
    for (j = 0; j < need; j++) printf "%d ", keys[j]
    # Anchor: one random empty-block height (no entry in jsonl)
    anchor = int(rand() * hi) + 1
    while (h[anchor]) anchor = int(rand() * hi) + 1
    printf "%d", anchor
  }' "$TXS_JSONL")"

echo
echo '  spot-check (last entry is a known-empty-block anchor):'
for h in $heights_csv; do
  rpc_num="$(rpc_block_header_field "$h" num_txs)"
  local_num="$(awk -v h="$h" -F'"' '
    /"block_height":/ {
      # Extract the numeric value after "block_height":"
      match($0, /"block_height":"[0-9]+"/)
      if (RSTART) {
        val = substr($0, RSTART, RLENGTH)
        gsub(/[^0-9]/, "", val)
        if (val == h) c++
      }
    }
    END { print c+0 }' "$TXS_JSONL")"

  if [[ "$rpc_num" == "$local_num" ]]; then
    printf '    height=%-8d rpc=%s  local=%s  [OK]\n' "$h" "$rpc_num" "$local_num"
  else
    printf '    height=%-8d rpc=%s  local=%s  [FAIL]\n' "$h" "$rpc_num" "$local_num"
    fail=1
  fi
done

echo
if [[ $fail -eq 0 ]]; then
  echo '  Result: all checks passed'
else
  echo '  Result: divergence detected — do not ship this txs.jsonl'
fi
exit "$fail"
