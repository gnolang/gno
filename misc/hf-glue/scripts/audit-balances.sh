#!/usr/bin/env bash
# audit-balances.sh — compare every historical tx signer's balance between
# the source chain at halt_height and our replay, to surface accounts that
# diverged (mostly: accounts with balance on mainnet that evaporated during
# replay because of post-mainnet storage-deposit semantics).
#
# Output
# ======
#   out/BALANCE-AUDIT.md — markdown table of every diverged signer with
#   mainnet balance, replay balance, and the ugnot delta. Rows sorted by
#   descending delta so the largest divergences appear first. The existing
#   hf_topup_balance pipeline in migrate.sh consumes a flat address list;
#   this audit's output is the raw data for choosing which addresses to
#   add to that pipeline.
#
# Why
# ===
#   Under --skip-failing-genesis-txs we observe ~2580 InsufficientFunds
#   failures at replay. Most come from historical signers whose balance on
#   mainnet was intact right up to halt_height, but which get drained in
#   replay by new storage-deposit charges that didn't exist on the source
#   chain. Without this audit, a production launch silently accepts those
#   balance drops — user txs that once worked would keep working if we
#   compensated.
#
# Approach
# ========
#   1. Walk out/source/txs.jsonl, extract (signer_address, tx_count). Pull
#      signer_info[0].address from the tx metadata; tx-archive produces one
#      address per tx, multi-sig not currently modeled here.
#   2. For each unique signer: query auth/accounts at mainnet@halt_height
#      and at the replay node. Parse coins → ugnot amount.
#   3. Emit the diff. No heuristics on criticality — that's a human call.
#
# Env
# ===
#   SOURCE_RPC   source chain RPC (default https://rpc.gno.land)
#   REPLAY_RPC   replay node RPC (default http://localhost:26657)
#   HALT_HEIGHT  source-chain height to query balances at (required)
#   TXS_JSONL    path to cached txs.jsonl (default $OUT/source/txs.jsonl)
#   OUT          misc/hf-glue/out (auto-resolved)
#   SIGNER_LIMIT unique signers to audit (default 200; use 0 for all).
#                Auditing all can be thousands of RPC calls.
#   GNOKEY_BIN   gnokey binary (default: gnokey on $PATH)
#
# Exit status
# ===========
#   0  — audit completed (even if divergences found — this is informational)
#   2  — prerequisite error
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="${OUT:-$(cd "$SCRIPT_DIR/.." && pwd)/out}"
mkdir -p "$OUT"

SOURCE_RPC="${SOURCE_RPC:-https://rpc.gno.land}"
REPLAY_RPC="${REPLAY_RPC:-http://localhost:26657}"
TXS_JSONL="${TXS_JSONL:-$OUT/source/txs.jsonl}"
SIGNER_LIMIT="${SIGNER_LIMIT:-200}"
GNOKEY_BIN="${GNOKEY_BIN:-gnokey}"

: "${HALT_HEIGHT:?HALT_HEIGHT is required}"

command -v "$GNOKEY_BIN" >/dev/null 2>&1 || {
  echo "gnokey not found" >&2
  exit 2
}
command -v jq >/dev/null 2>&1 || {
  echo "jq not found" >&2
  exit 2
}
[[ -f "$TXS_JSONL" ]] || {
  echo "txs.jsonl not found at $TXS_JSONL" >&2
  exit 2
}

REPORT="$OUT/BALANCE-AUDIT.md"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# ---- Helpers

# bank/balances/<addr> returns a one-line data response:
#   data: "<amt>ugnot,<amt>denom2,..."
# We parse only ugnot. Using bank/balances instead of auth/accounts because
# the latter multi-line JSON is annoying to parse from gnokey's text output.
query_balance_ugnot() {
  local rpc="$1" addr="$2" height="${3:-0}"
  local args=(-remote "$rpc")
  [[ "$height" -gt 0 ]] && args+=(-height "$height")

  local data
  data="$("$GNOKEY_BIN" query "${args[@]}" "bank/balances/$addr" 2>/dev/null |
    awk '/^data:/{sub(/^data: /,""); print; exit}')"
  # data is a quoted string: "754954090ugnot,123foo" (or "" for empty).
  # Strip surrounding quotes, then extract ugnot amount.
  data="${data#\"}"
  data="${data%\"}"
  if [[ -z "$data" ]]; then
    echo 0
    return
  fi
  local ugnot
  ugnot="$(printf '%s' "$data" |
    tr ',' '\n' |
    awk '/ugnot$/ {
      sub(/ugnot$/,"")
      print
      exit
    }')"
  echo "${ugnot:-0}"
}

# Extract unique signer addresses from txs.jsonl, newest-first by their
# occurrence order. BSD awk: no `length()` on arrays with gawk-style
# semantics, but we sort by count after.
extract_signers() {
  jq -r '.metadata.signer_info[0].address // empty' <"$TXS_JSONL" |
    sort | uniq -c | sort -rn |
    awk '{ print $2 "\t" $1 }'
}

# ---- Extract signer set
echo "━━━ audit-balances ━━━"
echo "  source  $SOURCE_RPC @ height=$HALT_HEIGHT"
echo "  replay  $REPLAY_RPC @ current tip"
echo "  txs     $TXS_JSONL"
echo "  limit   $SIGNER_LIMIT unique signers (0 = all)"
echo

extract_signers >"$WORK/signers.tsv"
total_signers="$(wc -l <"$WORK/signers.tsv" | tr -d ' ')"
echo "  found $total_signers unique signers in txs.jsonl"

if [[ "$SIGNER_LIMIT" -gt 0 ]]; then
  head -n "$SIGNER_LIMIT" "$WORK/signers.tsv" >"$WORK/audit.tsv"
else
  cp "$WORK/signers.tsv" "$WORK/audit.tsv"
fi
audit_count="$(wc -l <"$WORK/audit.tsv" | tr -d ' ')"
echo "  auditing $audit_count"
echo

# ---- Per-signer balance query
: >"$WORK/rows.tsv"
i=0
while IFS=$'\t' read -r addr tx_count; do
  i=$((i + 1))
  printf '  [%d/%d] %s ...\r' "$i" "$audit_count" "$addr"

  src_bal="$(query_balance_ugnot "$SOURCE_RPC" "$addr" "$HALT_HEIGHT")"
  rep_bal="$(query_balance_ugnot "$REPLAY_RPC" "$addr" 0)"
  delta=$((src_bal - rep_bal))

  printf '%s\t%s\t%s\t%s\t%s\n' "$addr" "$tx_count" "$src_bal" "$rep_bal" "$delta" \
    >>"$WORK/rows.tsv"
done <"$WORK/audit.tsv"
echo

# ---- Build report
# Sort by |delta| descending so the most diverged accounts are first.
sort -t$'\t' -k5 -rn "$WORK/rows.tsv" >"$WORK/rows.sorted.tsv"

{
  echo "# Balance audit"
  echo ""
  echo "_Generated $(date -u +%Y-%m-%dT%H:%M:%SZ)_"
  echo ""
  echo "- **Source**: \`$SOURCE_RPC\` at \`height=$HALT_HEIGHT\`"
  echo "- **Replay**: \`$REPLAY_RPC\` at current tip"
  echo "- **Audited**: $audit_count of $total_signers unique signers from \`txs.jsonl\`"
  echo ""
  echo "All amounts in ugnot. \`delta = source − replay\`; positive means the"
  echo "account had more on mainnet than we have post-replay — a candidate"
  echo "for \`hf_topup_balance\`. Zero delta means the replay matches source"
  echo "balance exactly."
  echo ""
  echo "| Address | tx count on source | mainnet @ halt | replay @ tip | delta |"
  echo "|---------|-------------------:|---------------:|-------------:|------:|"

  while IFS=$'\t' read -r addr tx_count src_bal rep_bal delta; do
    [[ -z "$addr" ]] && continue
    printf '| `%s` | %s | %s | %s | %s |\n' \
      "$addr" "$tx_count" "$src_bal" "$rep_bal" "$delta"
  done <"$WORK/rows.sorted.tsv"

  echo ""
  echo "---"
  echo ""
  echo "## Summary"
  diverged="$(awk -F'\t' '$5 != 0 { c++ } END { print c+0 }' "$WORK/rows.sorted.tsv")"
  positive="$(awk -F'\t' '$5 > 0 { c++ } END { print c+0 }' "$WORK/rows.sorted.tsv")"
  sum_positive="$(awk -F'\t' '$5 > 0 { s += $5 } END { print s+0 }' "$WORK/rows.sorted.tsv")"
  echo ""
  echo "- Accounts audited: **$audit_count**"
  echo "- Diverged (delta ≠ 0): **$diverged**"
  echo "- Replay short of source (delta > 0): **$positive** accounts, total **$sum_positive ugnot**"
  echo ""
  if [[ "$SIGNER_LIMIT" -gt 0 && "$total_signers" -gt "$audit_count" ]]; then
    echo "_Partial audit: $((total_signers - audit_count)) signer(s) not audited._"
    echo "_Set SIGNER_LIMIT=0 to audit all signers._"
    echo ""
  fi
} >"$REPORT"

echo "Report: $REPORT"
echo
printf 'Summary: audited=%d diverged=%d\n' "$audit_count" "$diverged"
