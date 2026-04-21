#!/usr/bin/env bash
# misc/hf-glue/scripts/migrate.sh
#
# Declarative hardfork migration — configured here, plumbed in lib/hf.sh.
# Defaults target gnoland1 → gnoland-1. Override by exporting any of
# SOURCE / RPC_URL / CHAIN_ID / ORIGINAL_CHAIN_ID / HALT_HEIGHT / PATCH_REALMS
# before running.
#
# Think of this file as a config that happens to be executable. Each hf_*
# call below is one line of intent; add / remove / reorder them to describe
# a different migration.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/hf.sh
source "$SCRIPT_DIR/lib/hf.sh"

hf_init

# -------------------------------------------------------------------------
# 1) Where to get the BASE GENESIS
# -------------------------------------------------------------------------
# Pick one. $SOURCE from the Makefile decides which branch runs.
: "${SOURCE:=https://github.com/gnolang/gno/releases/download/chain/gnoland1.0/genesis.json}"
case "$SOURCE" in
*.json | */genesis.json) hf_fetch_genesis_from_url "$SOURCE" ;;
http://* | https://*) hf_fetch_genesis_from_rpc "$SOURCE" ;;
*) hf_fetch_genesis_from_file "$SOURCE" ;;
esac

# -------------------------------------------------------------------------
# 2) Where to get the HISTORICAL TXS
# -------------------------------------------------------------------------
: "${RPC_URL:=https://rpc.gno.land}"
hf_fetch_txs_via_rpc "$RPC_URL"
# Alternatives:
#   hf_fetch_txs_from_jsonl /path/to/txs.jsonl
#   hf_skip_txs

# -------------------------------------------------------------------------
# 3) REALM PATCHES (ride along the hardfork)
# -------------------------------------------------------------------------
# Swap r/sys/params with the repo's current examples copy. After merging
# #5368 that copy has halt.gno (NewSetHaltRequest), so the forked chain
# boots with the govDAO halt mechanism available.
hf_patch_addpkg "gno.land/r/sys/params" "$REPO/examples/gno.land/r/sys/params"

# Extra patches from $PATCH_REALMS (space-separated PKGPATH=SRCDIR).
for spec in ${PATCH_REALMS:-}; do
  [[ -z "$spec" ]] && continue
  hf_patch_addpkg "${spec%%=*}" "${spec#*=}"
done

# -------------------------------------------------------------------------
# 3b) BALANCE TOP-UPS (pre-replay synthetic seeding)
# -------------------------------------------------------------------------
# Last-resort escape hatch when a genesis-mode tx would fail under newer
# code because an invariant that didn't exist when the tx was signed
# cannot be covered from the replay's starting state.
#
# This diverges the replayed state from the source chain — use only when
# the alternative is an incomplete replay. Every top-up is written to
# out/TOPUP-REPORT.md and surfaced in out/STATE-REPORT.md so the synthetic
# change is visible in the audit trail, not just in the fetch console.
#
# Current case (r/sys/txfees / tx_index=76):
#   Under master's storage-deposit logic (added after gnoland1 launched),
#   every addpkg *locks* a deposit from msg.Creator proportional to realm
#   size. The creator g1r929wt2qplfawe4lvqv9zuwfdcz4vxdun7qh8l deploys 7
#   r/sys/* realms in a row and exhausts its 58.7 Mugnot balance at the
#   8th (r/sys/txfees). Since all 85 genesis-mode addpkg txs were signed
#   with max_deposit="" (the field didn't exist on gnoland1) and the
#   realm code on master is identical to gnoland1's, no cherry-pick fixes
#   this — the gap is purely between old txs and new SDK semantics.
#   Top up the creator so the locks fit.
hf_topup_balance "g1r929wt2qplfawe4lvqv9zuwfdcz4vxdun7qh8l" "1000000000ugnot" \
  "storage-deposit headroom for r/sys/* genesis-mode deploys"

# -------------------------------------------------------------------------
# 4) OVERLAY TXS (pre-history, not yet supported)
# -------------------------------------------------------------------------
# Future: inject extra txs between genesis-mode and historical replay.
# hf_overlay_txs "$SCRIPT_DIR/../overlays/20260417_add_moderator.jsonl"

# -------------------------------------------------------------------------
# 5) MIGRATION TXS (post-history)
# -------------------------------------------------------------------------
# These run AFTER historical replay — "reproduce history, then mutate".
#
# Valset swap: gnoland1 seeds its valset via govdao_prop1.gno, so the
# post-fork r/sys/validators/v2 still lists the *original* 7 validators
# even though tm2 consensus is driven by GenesisDoc.Validators (which
# `gnogenesis fork` rewrites to our local validator via fixvalidator).
# The migration below reconciles the two: it wipes the 7 originals and
# registers the new valset via a govDAO proposal signed as manfred
# (T1 member) under --skip-genesis-sig-verification.
#
# Delegates to misc/deployments/gnoland-1/migrations/build.sh, which
# renders the template with the local priv_validator_key.json and
# produces a signed jsonl under $OUT/migrations.jsonl.
PV_KEY_DEFAULT="$OUT/gnoland-home/secrets/priv_validator_key.json"
PV_KEY="${PV_KEY:-$PV_KEY_DEFAULT}"
if [[ -f "$PV_KEY" ]]; then
  hf_banner "step 5 — post-replay migration (valset swap${NEW_T1_ADDR:+ + T1 rotation})"
  hf_kv "pv_key" "$PV_KEY"
  [[ -n "${NEW_T1_ADDR:-}" ]] && hf_kv "new T1 addr" "$NEW_T1_ADDR"
  MIG_JSONL="$OUT/migrations.jsonl"
  CALLER="${CALLER:-g1manfred47kzduec920z88wfr64ylksmdcedlf5}" \
    PV_KEY="$PV_KEY" \
    RPC_URL="$RPC_URL" \
    NEW_T1_ADDR="${NEW_T1_ADDR:-}" \
    T1_PORTFOLIO="${T1_PORTFOLIO:-}" \
    T1_WITHDRAW_REASON="${T1_WITHDRAW_REASON:-}" \
    OUT_JSONL="$MIG_JSONL" \
    CHAIN_ID="$CHAIN_ID" \
    REPO_ROOT="$REPO" \
    bash "$REPO/misc/deployments/gnoland-1/migrations/build.sh"
  hf_migration_tx "$MIG_JSONL"
else
  hf_banner "step 5 — post-replay migration (skipped)"
  hf_kv "reason" "no priv_validator_key.json at $PV_KEY — run 'make init' first"
fi

# -------------------------------------------------------------------------
# 6) ASSEMBLE the hardfork genesis
# -------------------------------------------------------------------------
hf_assemble
