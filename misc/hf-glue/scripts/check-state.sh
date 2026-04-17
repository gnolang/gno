#!/usr/bin/env bash
# Probe the running hardfork node, compare against live gno.land,
# write a STATE-REPORT.md with findings.
#
# Usage: ./scripts/check-state.sh [address]
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="$HERE/out"
REPORT="$OUT/STATE-REPORT.md"
mkdir -p "$OUT"

LOCAL_RPC="${LOCAL_RPC:-http://127.0.0.1:26657}"
LOCAL_WEB="${LOCAL_WEB:-http://127.0.0.1:8888}"
PROD_RPC="${PROD_RPC:-https://rpc.gno.land}"
PROD_WEB="${PROD_WEB:-https://gno.land}"
ADDRESS="${1:-g1manfred47kzduec920z88wfr64ylksmdcedlf5}"

b64() { printf '%s' "$1" | base64 | tr -d '\n'; }

# Query vm/qrender: prints one of
#   "OK:<first 80 chars of rendered output>"
#   "ERR:<error type>"
#   "UNREACHABLE"
qrender() {
  local rpc="$1" path="$2"
  local data resp
  data=$(b64 "$path")
  resp=$(curl -sS --max-time 10 "${rpc}/abci_query?path=%22vm%2Fqrender%22&data=${data}" 2>/dev/null || echo "")
  if [[ -z "$resp" ]]; then
    echo "UNREACHABLE"
    return
  fi
  local err
  err=$(printf '%s' "$resp" | jq -r '.result.response.ResponseBase.Error."@type" // empty' 2>/dev/null || echo "")
  if [[ -n "$err" ]]; then
    echo "ERR:$err"
    return
  fi
  local data_out
  data_out=$(printf '%s' "$resp" | jq -r '.result.response.ResponseBase.Data // empty' 2>/dev/null || echo "")
  local preview
  preview=$(printf '%s' "$data_out" | tr -d '\n' | head -c 80)
  echo "OK:${preview}"
}

# Query chain status as JSON
chain_status() {
  local rpc="$1"
  curl -sS --max-time 5 "${rpc}/status" 2>/dev/null \
    | jq -c '{chain_id: .result.node_info.network, latest_block: .result.sync_info.latest_block_height, catching_up: .result.sync_info.catching_up}' \
    2>/dev/null || echo '"unreachable"'
}

# Query account balance via auth/accounts
account_info() {
  local rpc="$1" addr="$2"
  local resp data
  resp=$(curl -sS --max-time 10 "${rpc}/abci_query?path=%22auth%2Faccounts%2F${addr}%22" 2>/dev/null || echo "")
  data=$(printf '%s' "$resp" | jq -r '.result.response.ResponseBase.Data // empty' 2>/dev/null || echo "")
  if [[ -z "$data" || "$data" == "null" ]]; then
    echo "(no account)"
    return
  fi
  printf '%s' "$data" | base64 -d 2>/dev/null | jq -c '.BaseAccount | {coins, account_number, sequence}' 2>/dev/null || echo "(decode failed)"
}

# ---- start report ----------------------------------------------------------
{
  echo "# Hardfork State Report"
  echo ""
  echo "_Generated $(date -u +%Y-%m-%dT%H:%M:%SZ)_"
  echo ""
  echo "- **Local**: $LOCAL_RPC (gnoweb: $LOCAL_WEB)"
  echo "- **Prod**:  $PROD_RPC  (gnoweb: $PROD_WEB)"
  echo "- **Address under test**: \`$ADDRESS\`"
  echo ""
  echo "## Chain status"
  echo ""
  echo "| Chain | Status |"
  echo "|-------|--------|"
  echo "| Local | \`$(chain_status "$LOCAL_RPC")\` |"
  echo "| Prod  | \`$(chain_status "$PROD_RPC")\` |"
  echo ""
  echo "## Expected realms (should exist after hardfork)"
  echo ""
  echo "| Realm | Local | Prod |"
  echo "|-------|-------|------|"
  for realm in \
    "gno.land/r/sys/params:" \
    "gno.land/r/sys/names:" \
    "gno.land/r/sys/users:" \
    "gno.land/r/gov/dao:" \
    "gno.land/r/gov/dao:proposals" \
    "gno.land/r/gnoland/home:" \
    "gno.land/r/gnoland/blog:" \
    "gno.land/r/gnoland/valopers:" \
    "gno.land/r/gnoland/coins:" \
    "gno.land/r/gnoland/wugnot:" \
  ; do
    l=$(qrender "$LOCAL_RPC" "$realm")
    p=$(qrender "$PROD_RPC"  "$realm")
    # collapse to status icon
    lt="❌ $l"; [[ $l == OK:* ]] && lt="✅"
    pt="❌ $p"; [[ $p == OK:* ]] && pt="✅"
    echo "| \`$realm\` | $lt | $pt |"
  done
  echo ""
  echo "## Bank balance — \`$ADDRESS\`"
  echo ""
  echo "| Chain | auth/accounts | r/gnoland/coins:balances |"
  echo "|-------|---------------|--------------------------|"
  la=$(account_info "$LOCAL_RPC" "$ADDRESS")
  pa=$(account_info "$PROD_RPC"  "$ADDRESS")
  lc=$(qrender "$LOCAL_RPC" "gno.land/r/gnoland/coins:balances?address=${ADDRESS}&coin" | head -c 120)
  pc=$(qrender "$PROD_RPC"  "gno.land/r/gnoland/coins:balances?address=${ADDRESS}&coin" | head -c 120)
  echo "| Local | \`$la\` | \`$lc\` |"
  echo "| Prod  | \`$pa\` | \`$pc\` |"
  echo ""
  echo "## Gas / consensus params"
  echo ""
  echo "### Local consensus"
  echo '```json'
  curl -sS --max-time 5 "$LOCAL_RPC/consensus_params" 2>/dev/null \
    | jq '.result.consensus_params.Block' 2>/dev/null || echo "(unreachable)"
  echo '```'
  echo ""
  echo "### Prod consensus"
  echo '```json'
  curl -sS --max-time 5 "$PROD_RPC/consensus_params" 2>/dev/null \
    | jq '.result.consensus_params.Block' 2>/dev/null || echo "(unreachable)"
  echo '```'
  echo ""
  echo "### Local auth params"
  echo '```json'
  curl -sS --max-time 10 "$LOCAL_RPC/abci_query?path=%22auth%2Fparams%22" 2>/dev/null \
    | jq -r '.result.response.ResponseBase.Data // empty' \
    | base64 -d 2>/dev/null \
    | jq '.' 2>/dev/null || echo "(no data)"
  echo '```'
  echo ""
  echo "### Prod auth params"
  echo '```json'
  curl -sS --max-time 10 "$PROD_RPC/abci_query?path=%22auth%2Fparams%22" 2>/dev/null \
    | jq -r '.result.response.ResponseBase.Data // empty' \
    | base64 -d 2>/dev/null \
    | jq '.' 2>/dev/null || echo "(no data)"
  echo '```'
  echo ""
  echo "## Visual comparison (open these side-by-side)"
  echo ""
  echo "| Page | Local | Prod |"
  echo "|------|-------|------|"
  for p in "r/gov/dao" "r/gnoland/blog" "r/gnoland/home" "r/sys/params" "r/gnoland/coins:balances?address=${ADDRESS}&coin"; do
    echo "| \`${p}\` | [$LOCAL_WEB/$p]($LOCAL_WEB/$p) | [$PROD_WEB/$p]($PROD_WEB/$p) |"
  done
  echo ""
} > "$REPORT"

echo "Report written to: $REPORT"
echo ""
cat "$REPORT"
