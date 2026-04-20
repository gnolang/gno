#!/usr/bin/env bash
# Synthesise the gnoland-1 hardfork migration jsonl from templates.
#
# Output: $OUT_JSONL (one amino-JSON TxWithMetadata per line).
# Passed to `gnogenesis fork generate --migration-tx $OUT_JSONL`.
#
# What it does
# ============
# 1. Fills 01_reset_valset.gno.tmpl with:
#      - OLD_VALIDATORS_GO   = voting_power=0 entries for all INITIAL_VALSET
#                              entries of gnoland1 (removes them from
#                              r/sys/validators/v2)
#      - NEW_VALIDATORS_GO   = the single post-fork validator described by
#                              $NEW_VALSET_JSON (produced by hf-glue
#                              init-node.sh, or by a manual list for a
#                              coordinated hardfork)
# 2. Wraps the rendered .gno body in a MsgRun tx signed by any local key;
#    the tx's `caller` field is set to $CALLER (a govDAO T1 member, e.g.
#    g1manfred...) so the proposal executes as that member when
#    --skip-genesis-sig-verification kicks in at replay.
# 3. Emits one `{tx: {...}}` per migration into $OUT_JSONL.
#
# Env
# ===
#   CALLER            govDAO T1 member address (required)
#                     default: g1manfred47kzduec920z88wfr64ylksmdcedlf5
#   NEW_VALSET_JSON   path to JSON with new validators, format:
#                       [{"address": "g1...", "pub_key": "gpub1...",
#                         "voting_power": 10, "name": "hf-local"}, ...]
#                     default: synthesised from a priv_validator_key.json if
#                     $PV_KEY is set.
#   PV_KEY            alternate: path to a priv_validator_key.json; if set
#                     and $NEW_VALSET_JSON is empty, a single-validator set
#                     is derived from it (power=10, name=hf-local).
#   OUT_JSONL         output path (default: ./migrations.jsonl)
#   GNOKEY_BIN        gnokey binary (auto-built if missing)
#   REPO_ROOT         repo root (auto-detected)
#
# Example
# =======
#   CALLER=g1manfred47kzduec920z88wfr64ylksmdcedlf5 \
#   PV_KEY=/path/to/priv_validator_key.json \
#   ./build.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="${REPO_ROOT:-$(cd "$SCRIPT_DIR/../../../.." && pwd)}"
OUT_JSONL="${OUT_JSONL:-$SCRIPT_DIR/migrations.jsonl}"
CALLER="${CALLER:-g1manfred47kzduec920z88wfr64ylksmdcedlf5}"
CHAIN_ID="${CHAIN_ID:-gnoland-1}"

# Initial gnoland1 valset (mirrors INITIAL_VALSET in
# misc/deployments/gnoland1/gen-genesis.sh). All seven are removed by this
# migration — add to this list if the source chain's valset changed post-
# genesis and should also be removed.
OLD_ADDRS=(
  "g1vta7dwp4guuhkfzksenfcheky4xf9hue8mgne4"
  "g1d5hh9fw3l00gugfzafskaxqlmsyvxfaj6l2q60"
  "g1uhv7wr7nku89se3t7v8fpquc7n5sf8rfkywxpc"
  "g10jdd8vlgydfypynrk23ul90jnsg5twrtvmcmh4"
  "g1eueypc9w524ctda3y0kwd4jruw5p4zqpjna0jq"
  "g1kn7p0wqumvqlcqzhkwnavkhf0z4qnr73ltwsae"
  "g10j90aqjv6uju3dksq8m08s6u47x59glkdxqzm2"
)

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
  [[ "$BECH_PUBKEY" == gpub1* ]] || { echo "ERROR: failed to derive bech32 pubkey from $PV_KEY (got: $BECH_PUBKEY)" >&2; exit 1; }
  ADDR="$(jq -r '.address' "$PV_KEY")"
  NEW_VALSET_JSON="$WORK/new_valset.json"
  jq -n --arg addr "$ADDR" --arg pub "$BECH_PUBKEY" '[{
    address:      $addr,
    pub_key:      $pub,
    voting_power: 10,
    name:         "hf-local"
  }]' > "$NEW_VALSET_JSON"
fi

# ---- render template ----
OLD_GO=""
for a in "${OLD_ADDRS[@]}"; do
  OLD_GO+="{Address: \"$a\", VotingPower: 0},"$'\n\t\t\t\t'
done

NEW_GO=$(jq -r '.[] | "{Address: \"\(.address)\", PubKey: \"\(.pub_key)\", VotingPower: \(.voting_power)},"' "$NEW_VALSET_JSON" | awk 'BEGIN{ORS="\n\t\t\t\t"}{print}')

RENDERED="$WORK/01_reset_valset.gno"
# awk-based substitution (BSD sed can't handle newlines in replacement).
OLD_GO="$OLD_GO" NEW_GO="$NEW_GO" awk '
  { gsub(/\{\{OLD_VALIDATORS_GO\}\}/, ENVIRON["OLD_GO"])
    gsub(/\{\{NEW_VALIDATORS_GO\}\}/, ENVIRON["NEW_GO"])
    print }
' "$SCRIPT_DIR/01_reset_valset.gno.tmpl" > "$RENDERED"

# ---- build the MsgRun tx ----
# We use any local ephemeral key to sign; only the serialized form matters
# because --skip-genesis-sig-verification is on at replay. The msg's
# `caller` field is what the VM reads (runtime.OriginCaller), so we patch
# it to $CALLER after maketx.
GK_HOME="$WORK/gnokey-home"
mkdir -p "$GK_HOME"
EPHEMERAL_MNEMONIC="source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
# stdin order for `gnokey add --recover --insecure-password-stdin`:
#   1. passphrase (empty line = no passphrase, no confirm prompt)
#   2. mnemonic
printf '\n%s\n' "$EPHEMERAL_MNEMONIC" | \
  "$GNOKEY_BIN" add --recover --insecure-password-stdin --home "$GK_HOME" ephemeral >/dev/null

TX_JSON="$WORK/tx.json"
"$GNOKEY_BIN" maketx run \
  --gas-wanted 100000000 \
  --gas-fee 1ugnot \
  --chainid "$CHAIN_ID" \
  --home "$GK_HOME" \
  ephemeral \
  "$RENDERED" > "$TX_JSON"

# Patch the caller field so the MsgRun executes as $CALLER (not ephemeral).
jq --arg caller "$CALLER" '.msg[0].caller = $caller' "$TX_JSON" > "$TX_JSON.patched"
mv "$TX_JSON.patched" "$TX_JSON"

# Sign (bogus sig, skipped at replay — but the tx format requires one).
echo "" | "$GNOKEY_BIN" sign \
  --tx-path "$TX_JSON" \
  --chainid "$CHAIN_ID" \
  --account-number 0 \
  --account-sequence 0 \
  --home "$GK_HOME" \
  --insecure-password-stdin \
  ephemeral >/dev/null

# Wrap as {tx: {...}} — TxWithMetadata accepts this with empty metadata;
# BlockHeight is forced to 0 by gnogenesis readMigrationTxs.
jq -c '{tx: .}' "$TX_JSON" > "$OUT_JSONL"

printf '  migration: %s (caller=%s)\n' "$(basename "$RENDERED")" "$CALLER"
printf '  written:   %s\n' "$OUT_JSONL"
