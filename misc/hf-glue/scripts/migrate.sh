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
  *.json|*/genesis.json)          hf_fetch_genesis_from_url  "$SOURCE" ;;
  http://*|https://*)             hf_fetch_genesis_from_rpc  "$SOURCE" ;;
  *)                              hf_fetch_genesis_from_file "$SOURCE" ;;
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
# 4) OVERLAY TXS (pre-history, not yet supported)
# -------------------------------------------------------------------------
# Future: inject extra txs between genesis-mode and historical replay.
# hf_overlay_txs "$SCRIPT_DIR/../overlays/20260417_add_moderator.jsonl"

# -------------------------------------------------------------------------
# 5) MIGRATION TXS (post-history, not yet supported)
# -------------------------------------------------------------------------
# Future: txs that run AFTER historical replay — "reproduce history,
# then mutate". Critical for valset swap: gnoland1 seeds its valset
# via govdao_prop1.gno, so a hardfork inherits the *original* 7
# validators inside r/sys/validators/v2 even though tm2 consensus is
# driven by GenesisDoc.Validators (which fixvalidator rewrites to our
# single local key). To reconcile the two, a migration tx should call
# r/sys/validators/v2.NewPropRequest via the govDAO flow to register
# the new valset, signed by a T1 member (manfred, under
# --skip-genesis-sig-verification).
#
# hf_migration_tx "$SCRIPT_DIR/../migrations/set_valset.msgrun.json"

# -------------------------------------------------------------------------
# 6) ASSEMBLE the hardfork genesis
# -------------------------------------------------------------------------
hf_assemble
