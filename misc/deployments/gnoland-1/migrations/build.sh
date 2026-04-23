#!/usr/bin/env bash
# Synthesise the gnoland-1 hardfork migration jsonl from templates.
#
# Output: $OUT_JSONL (one amino-JSON TxWithMetadata per line).
# Passed to `gnogenesis fork generate --migration-tx $OUT_JSONL`.
#
# What it does
# ============
# 1. Valset reset — fills 01_reset_valset.gno.tmpl with:
#      - OLD_VALIDATORS_GO   = voting_power=0 entries for all INITIAL_VALSET
#                              entries of gnoland1 (removes them from
#                              r/sys/validators/v2)
#      - NEW_VALIDATORS_GO   = the single post-fork validator described by
#                              $NEW_VALSET_JSON (produced by hf-glue
#                              init-node.sh, or by a manual list for a
#                              coordinated hardfork)
#    Signed by $CALLER (sole T1, manfred).
#
# 2. T1 rotation (optional, enabled when $NEW_T1_ADDR is set). 3 additional
#    txs in the jsonl:
#      - 02 AddMember(NEW_T1_ADDR, T1)  — manfred proposes, votes, executes
#                                          (100% supermajority as sole T1)
#      - 03 WithdrawMember(manfred)     — manfred proposes + votes YES
#                                          (50% of 2 T1s, not executed yet)
#      - 04 (caller=NEW_T1_ADDR) finds the open Withdraw proposal, votes YES,
#           executes (100% with both voting YES)
#
# 3. Wraps each rendered .gno body in a MsgRun tx signed by a local ephemeral
#    key; the tx's `caller` field is patched to the appropriate T1 member so
#    --skip-genesis-sig-verification at replay uses that as OriginCaller.
# 4. Emits one `{tx: {...}}` per migration into $OUT_JSONL.
#
# Env
# ===
#   CALLER              govDAO T1 member address for the valset-reset + add-member
#                       + withdraw-propose txs (required)
#                       default: g1manfred47kzduec920z88wfr64ylksmdcedlf5
#   RPC_URL             source-chain RPC (required). Queried once via
#                       abci_query vm/qeval to derive OLD_ADDRS from the
#                       *current* r/sys/validators/v2 state, so the valset
#                       reset only attempts to remove validators that actually
#                       exist at fork time. When unset, falls back to the
#                       hardcoded gnoland1 INITIAL_VALSET below (may be stale
#                       — will panic at replay if the source chain removed any
#                       of them via historical govDAO proposals).
#                       default: (unset — uses hardcoded fallback)
#   NEW_T1_ADDR         address to install as the sole T1 member of the
#                       post-fork govDAO. When set, three extra migration txs
#                       are appended (see 2. above). When empty, only the
#                       valset reset is emitted.
#   T1_PORTFOLIO        human-readable portfolio/justification attached to the
#                       AddMember proposal (required when NEW_T1_ADDR is set).
#                       default: "post-hardfork T1 rotation"
#   T1_WITHDRAW_REASON  reason string attached to the WithdrawMember proposal
#                       (r/gov/dao requires this for T1 removals).
#                       default: "replaced by NEW_T1_ADDR as part of hardfork"
#   NEW_VALSET_JSON     path to JSON with new validators, format:
#                         [{"address": "g1...", "pub_key": "gpub1...",
#                           "voting_power": 10, "name": "hf-local"}, ...]
#                       default: synthesised from a priv_validator_key.json if
#                       $PV_KEY is set.
#   PV_KEY              alternate: path to a priv_validator_key.json; if set
#                       and $NEW_VALSET_JSON is empty, a single-validator set
#                       is derived from it (power=10, name=hf-local).
#   OUT_JSONL           output path (default: ./migrations.jsonl)
#   GNOKEY_BIN          gnokey binary (auto-built if missing)
#   REPO_ROOT           repo root (auto-detected)
#
# Example (valset-only)
# =====================
#   CALLER=g1manfred47kzduec920z88wfr64ylksmdcedlf5 \
#   PV_KEY=/path/to/priv_validator_key.json \
#   ./build.sh
#
# Example (valset + T1 rotation)
# ==============================
#   CALLER=g1manfred47kzduec920z88wfr64ylksmdcedlf5 \
#   PV_KEY=/path/to/priv_validator_key.json \
#   NEW_T1_ADDR=g1yournewcontrolleraddresshere \
#   T1_PORTFOLIO="Core dev, handing over from moul" \
#   ./build.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="${REPO_ROOT:-$(cd "$SCRIPT_DIR/../../../.." && pwd)}"
OUT_JSONL="${OUT_JSONL:-$SCRIPT_DIR/migrations.jsonl}"
CALLER="${CALLER:-g1manfred47kzduec920z88wfr64ylksmdcedlf5}"
CHAIN_ID="${CHAIN_ID:-gnoland-1}"

NEW_T1_ADDR="${NEW_T1_ADDR:-}"
T1_PORTFOLIO="${T1_PORTFOLIO:-post-hardfork T1 rotation}"
T1_WITHDRAW_REASON="${T1_WITHDRAW_REASON:-replaced by NEW_T1_ADDR as part of hardfork}"

# Hardcoded fallback: gnoland1 INITIAL_VALSET (mirrors misc/deployments/
# gnoland1/gen-genesis.sh). Used only when $RPC_URL is unset. Likely stale at
# any real fork point because govDAO proposals may have added/removed valset
# entries post-genesis — prefer RPC-derived OLD_ADDRS whenever possible.
FALLBACK_OLD_ADDRS=(
  "g1vta7dwp4guuhkfzksenfcheky4xf9hue8mgne4"
  "g1d5hh9fw3l00gugfzafskaxqlmsyvxfaj6l2q60"
  "g1uhv7wr7nku89se3t7v8fpquc7n5sf8rfkywxpc"
  "g10jdd8vlgydfypynrk23ul90jnsg5twrtvmcmh4"
  "g1eueypc9w524ctda3y0kwd4jruw5p4zqpjna0jq"
  "g1kn7p0wqumvqlcqzhkwnavkhf0z4qnr73ltwsae"
  "g10j90aqjv6uju3dksq8m08s6u47x59glkdxqzm2"
)

# ---- resolve OLD_ADDRS from live r/sys/validators/v2, or fall back ----
# abci_query vm/qeval returns the amino-printed slice of
# gno.land/p/sys/validators.Validator values. We don't parse the full amino
# pretty-print — we just regex-extract every bech32 g1 address from the
# decoded payload. This is safe because the bech32 data-charset for gpub1
# pubkeys excludes the digit `1`, so the sequence `g1` only occurs at the
# start of address values, never inside a pubkey.
query_current_valset_addrs() {
  local rpc="$1"
  local data_b64 resp data
  data_b64=$(printf '%s' 'gno.land/r/sys/validators/v2.GetValidators()' | openssl base64 -A)
  resp=$(curl -fsS -X POST -H 'Content-Type: application/json' \
    -d "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"abci_query\",\"params\":{\"path\":\"vm/qeval\",\"data\":\"$data_b64\"}}" \
    "$rpc") || return 1
  data=$(jq -r '.result.response.ResponseBase.Data // empty' <<<"$resp")
  [[ -n "$data" ]] || return 1
  printf '%s' "$data" | openssl base64 -d -A | grep -oE 'g1[0-9a-z]{38}'
}

OLD_ADDRS=()
if [[ -n "${RPC_URL:-}" ]]; then
  while IFS= read -r addr; do
    [[ -n "$addr" ]] && OLD_ADDRS+=("$addr")
  done < <(query_current_valset_addrs "$RPC_URL")
  if [[ ${#OLD_ADDRS[@]} -eq 0 ]]; then
    echo "ERROR: failed to derive OLD_ADDRS from RPC $RPC_URL (empty response or no g1 addresses)" >&2
    exit 1
  fi
  echo "  valset source: RPC $RPC_URL (${#OLD_ADDRS[@]} validator(s))"
else
  OLD_ADDRS=("${FALLBACK_OLD_ADDRS[@]}")
  echo "  valset source: FALLBACK (hardcoded INITIAL_VALSET, ${#OLD_ADDRS[@]} validator(s))"
  echo "  WARNING: no RPC_URL set — valset reset may panic if source chain removed any of these." >&2
fi

# ---- resolve GNOKEY_BIN ----
if [[ -z "${GNOKEY_BIN:-}" ]]; then
  if command -v gnokey >/dev/null 2>&1; then
    GNOKEY_BIN="gnokey"
  else
    GNOKEY_BIN="$REPO_ROOT/contribs/gnogenesis/genesis-work/bin/gnokey"
    [[ -x "$GNOKEY_BIN" ]] || {
      mkdir -p "$(dirname "$GNOKEY_BIN")"
      go build -C "$REPO_ROOT/gno.land/cmd/gnokey" -o "$GNOKEY_BIN" .
    }
  fi
fi

# ---- assemble NEW_VALSET_JSON from PV_KEY if needed ----
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

if [[ -z "${NEW_VALSET_JSON:-}" ]]; then
  : "${PV_KEY:?either NEW_VALSET_JSON or PV_KEY is required}"
  # r/sys/validators/v2 wants the bech32 (gpub1...) pubkey — priv_validator_key.json
  # stores the raw base64 under pub_key.value. Use `gnoland secrets get` to convert.
  SECRETS_DIR="$(dirname "$PV_KEY")"
  BECH_PUBKEY="$(go run -C "$REPO_ROOT" ./gno.land/cmd/gnoland secrets get validator_key.pub_key --raw -data-dir "$SECRETS_DIR" | tail -n 1 | tr -d '[:space:]')"
  [[ "$BECH_PUBKEY" == gpub1* ]] || {
    echo "ERROR: failed to derive bech32 pubkey from $PV_KEY (got: $BECH_PUBKEY)" >&2
    exit 1
  }
  ADDR="$(jq -r '.address' "$PV_KEY")"
  NEW_VALSET_JSON="$WORK/new_valset.json"
  jq -n --arg addr "$ADDR" --arg pub "$BECH_PUBKEY" '[{
    address:      $addr,
    pub_key:      $pub,
    voting_power: 10,
    name:         "hf-local"
  }]' >"$NEW_VALSET_JSON"
fi

# ---- set up ephemeral signing key (used for all migration txs) ----
GK_HOME="$WORK/gnokey-home"
mkdir -p "$GK_HOME"
EPHEMERAL_MNEMONIC="source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
# stdin order for `gnokey add --recover --insecure-password-stdin`:
#   1. mnemonic
#   2. passphrase (empty line = no passphrase, skips confirm prompt)
printf '%s\n\n' "$EPHEMERAL_MNEMONIC" |
  "$GNOKEY_BIN" add --recover --insecure-password-stdin --home "$GK_HOME" ephemeral >/dev/null

# ---- helper: render .gno file from template, wrap into a signed tx, patch caller ----
# Args: <rendered_gno_path> <tx_caller_addr>
# Prints the resulting {tx: {...}} line on stdout.
render_tx() {
  local gno_path="$1" caller="$2"
  local tx_json="$WORK/$(basename "$gno_path" .gno).tx.json"

  "$GNOKEY_BIN" maketx run \
    --gas-wanted 100000000 \
    --gas-fee 1ugnot \
    --chainid "$CHAIN_ID" \
    --home "$GK_HOME" \
    ephemeral \
    "$gno_path" >"$tx_json"

  # Patch the caller field so the MsgRun executes as $caller (not ephemeral).
  jq --arg caller "$caller" '.msg[0].caller = $caller' "$tx_json" >"$tx_json.patched"
  mv "$tx_json.patched" "$tx_json"

  # Sign (bogus sig, skipped at replay — but the tx format requires one).
  echo "" | "$GNOKEY_BIN" sign \
    --tx-path "$tx_json" \
    --chainid "$CHAIN_ID" \
    --account-number 0 \
    --account-sequence 0 \
    --home "$GK_HOME" \
    --insecure-password-stdin \
    ephemeral >/dev/null

  # Wrap as {tx: {...}} — TxWithMetadata accepts this with empty metadata;
  # BlockHeight is forced to 0 by gnogenesis readMigrationTxs.
  jq -c '{tx: .}' "$tx_json"
}

# ---- helper: render a template with placeholder=value pairs ----
# Args: <template_path> <output_path> <PLACEHOLDER1=VALUE1> [<PLACEHOLDER2=VALUE2> ...]
render_template() {
  local tmpl="$1" out="$2"
  shift 2
  cp "$tmpl" "$out"
  local spec name val
  for spec in "$@"; do
    name="${spec%%=*}"
    val="${spec#*=}"
    # awk-based substitution (BSD sed can't reliably handle newlines + special chars in replacement).
    PH_NAME="$name" PH_VAL="$val" awk '
      BEGIN { name = ENVIRON["PH_NAME"]; val = ENVIRON["PH_VAL"] }
      { gsub("\\{\\{" name "\\}\\}", val); print }
    ' "$out" >"$out.tmp"
    mv "$out.tmp" "$out"
  done
}

# ---- 1. valset reset tx (caller=manfred) ----
OLD_GO=""
for a in "${OLD_ADDRS[@]}"; do
  OLD_GO+="{Address: \"$a\", VotingPower: 0},"$'\n\t\t\t\t'
done

NEW_GO=$(jq -r '.[] | "{Address: \"\(.address)\", PubKey: \"\(.pub_key)\", VotingPower: \(.voting_power)},"' "$NEW_VALSET_JSON" | awk 'BEGIN{ORS="\n\t\t\t\t"}{print}')

RENDERED_01="$WORK/01_reset_valset.gno"
render_template "$SCRIPT_DIR/01_reset_valset.gno.tmpl" "$RENDERED_01" \
  "OLD_VALIDATORS_GO=$OLD_GO" "NEW_VALIDATORS_GO=$NEW_GO"

: >"$OUT_JSONL"
render_tx "$RENDERED_01" "$CALLER" >>"$OUT_JSONL"
printf '  migration: %-38s caller=%s\n' "$(basename "$RENDERED_01")" "$CALLER"

# ---- 2-4. T1 rotation (optional) ----
if [[ -n "$NEW_T1_ADDR" ]]; then
  # Basic sanity checks on NEW_T1_ADDR.
  [[ "$NEW_T1_ADDR" =~ ^g1[0-9a-z]{38}$ ]] || {
    echo "ERROR: NEW_T1_ADDR does not look like a valid bech32 address: $NEW_T1_ADDR" >&2
    exit 1
  }
  [[ "$NEW_T1_ADDR" != "$CALLER" ]] || {
    echo "ERROR: NEW_T1_ADDR must differ from CALLER (no-op rotation)" >&2
    exit 1
  }

  # 02 — manfred adds NEW_T1_ADDR as T1 (100% supermajority, passes).
  RENDERED_02="$WORK/02_add_t1_member.gno"
  render_template "$SCRIPT_DIR/02_add_t1_member.gno.tmpl" "$RENDERED_02" \
    "NEW_T1_ADDR=$NEW_T1_ADDR" "PORTFOLIO=$T1_PORTFOLIO"
  render_tx "$RENDERED_02" "$CALLER" >>"$OUT_JSONL"
  printf '  migration: %-38s caller=%s\n' "$(basename "$RENDERED_02")" "$CALLER"

  # 03 — manfred proposes his own withdrawal + votes YES (50%, not executed).
  RENDERED_03="$WORK/03_withdraw_manfred_propose.gno"
  render_template "$SCRIPT_DIR/03_withdraw_manfred_propose.gno.tmpl" "$RENDERED_03" \
    "OLD_T1_ADDR=$CALLER" "WITHDRAW_REASON=$T1_WITHDRAW_REASON"
  render_tx "$RENDERED_03" "$CALLER" >>"$OUT_JSONL"
  printf '  migration: %-38s caller=%s\n' "$(basename "$RENDERED_03")" "$CALLER"

  # 04 — NEW_T1_ADDR votes YES on the open withdraw prop + executes.
  RENDERED_04="$WORK/04_withdraw_manfred_execute.gno"
  render_template "$SCRIPT_DIR/04_withdraw_manfred_execute.gno.tmpl" "$RENDERED_04" \
    "OLD_T1_ADDR=$CALLER"
  render_tx "$RENDERED_04" "$NEW_T1_ADDR" >>"$OUT_JSONL"
  printf '  migration: %-38s caller=%s\n' "$(basename "$RENDERED_04")" "$NEW_T1_ADDR"
fi

# ---- 5-7. deploy r/sys/validators/v3 ----
# v3 introduces the params-keeper-driven valset flow (see PR #5485). Mainnet
# never had it, so a fresh addpkg is needed post-fork. gnoland1's r/sys/names
# namespace check is enabled at halt height, so a direct addpkg under
# r/sys/* returns "unauthorized user". Strategy: wrap the addpkg with a
# temporary VM-param flip — steps 05/07 disable/restore the namespace check
# via govDAO proposals (the only authorized path to set vm:p:sysnames_pkgpath
# is `r/sys/params.NewSysParamStringPropRequest`).
V3_PKGDIR="${V3_PKGDIR:-$REPO_ROOT/examples/gno.land/r/sys/validators/v3}"
[[ -d "$V3_PKGDIR" ]] || {
  echo "ERROR: v3 pkgdir not found: $V3_PKGDIR" >&2
  exit 1
}

# Current sole T1 member at this point in the migration sequence. If T1
# rotation ran (steps 02-04), manfred is no longer T1; NEW_T1_ADDR is. The
# govDAO proposals in 05/07 need supermajority from the current T1.
if [[ -n "$NEW_T1_ADDR" ]]; then
  T1_CALLER="$NEW_T1_ADDR"
else
  T1_CALLER="$CALLER"
fi

# 05 — disable namespace check (govDAO proposal, caller=T1_CALLER).
RENDERED_05="$WORK/05_disable_sysnames.gno"
cp "$SCRIPT_DIR/05_disable_sysnames.gno.tmpl" "$RENDERED_05"
render_tx "$RENDERED_05" "$T1_CALLER" >>"$OUT_JSONL"
printf '  migration: %-38s caller=%s\n' "$(basename "$RENDERED_05")" "$T1_CALLER"

# 06 — addpkg r/sys/validators/v3 (MsgAddPackage, creator=manfred; sig-skip
# applies since this is a genesis-mode migration tx).
RENDERED_06="$WORK/06_addpkg_validators_v3.tx.json"
"$GNOKEY_BIN" maketx addpkg \
  --gas-wanted 100000000 \
  --gas-fee 1ugnot \
  --pkgpath "gno.land/r/sys/validators/v3" \
  --pkgdir "$V3_PKGDIR" \
  --chainid "$CHAIN_ID" \
  --home "$GK_HOME" \
  ephemeral >"$RENDERED_06"

# MsgAddPackage uses `creator` (not `caller`). Patch to manfred so the
# addpkg runs as him under --skip-genesis-sig-verification.
jq --arg creator "$CALLER" '.msg[0].creator = $creator' "$RENDERED_06" >"$RENDERED_06.patched"
mv "$RENDERED_06.patched" "$RENDERED_06"

echo "" | "$GNOKEY_BIN" sign \
  --tx-path "$RENDERED_06" \
  --chainid "$CHAIN_ID" \
  --account-number 0 \
  --account-sequence 0 \
  --home "$GK_HOME" \
  --insecure-password-stdin \
  ephemeral >/dev/null

jq -c '{tx: .}' "$RENDERED_06" >>"$OUT_JSONL"
printf '  migration: %-38s caller=%s\n' "06_addpkg_validators_v3" "$CALLER"

# 07 — restore namespace check (govDAO proposal, caller=T1_CALLER).
RENDERED_07="$WORK/07_restore_sysnames.gno"
cp "$SCRIPT_DIR/07_restore_sysnames.gno.tmpl" "$RENDERED_07"
render_tx "$RENDERED_07" "$T1_CALLER" >>"$OUT_JSONL"
printf '  migration: %-38s caller=%s\n' "$(basename "$RENDERED_07")" "$T1_CALLER"

# 08 — point vm:p:valset_realm_path at the v3 realm (govDAO proposal,
# caller=T1_CALLER). Without this, EndBlocker reads valsetRealm="" from
# pre-v3 mainnet state and never picks up the updates r/sys/validators/v3
# writes to its params.
RENDERED_08="$WORK/08_set_valset_realm.gno"
cp "$SCRIPT_DIR/08_set_valset_realm.gno.tmpl" "$RENDERED_08"
render_tx "$RENDERED_08" "$T1_CALLER" >>"$OUT_JSONL"
printf '  migration: %-38s caller=%s\n' "$(basename "$RENDERED_08")" "$T1_CALLER"

printf '  written:   %s\n' "$OUT_JSONL"
