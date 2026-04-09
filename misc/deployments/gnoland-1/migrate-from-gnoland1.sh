#!/usr/bin/env bash
# migrate-from-gnoland1.sh — hard fork gnoland1 → gnoland-1.
#
# This script produces a new genesis.json for the gnoland-1 chain from the
# final committed state of gnoland1 after it halts at the governance-approved
# halt height.
#
# Approach: Scenario A (genesis tx-replay with InitialHeight preservation)
#   - Export all successful txs from gnoland1 with full block metadata
#     (height, timestamp, chain_id) via tx-archive.
#   - Assemble a new genesis.json for gnoland-1 using genesis-assemble:
#       chain_id: "gnoland-1"
#       initial_height: <halt block height + 1>
#       app_state.txs: full tx array with metadata preserved
#   - Height AND timestamp are preserved (Jae's correctness requirement).
#   - Validators independently run this script and compare genesis SHA-256
#     before restarting.
#
# Usage:
#   ./migrate-from-gnoland1.sh --data-dir <gnoland1-data-dir> [--halt-height N]
#
# The script writes genesis.json (gnoland-1) to the current directory.
#
# Prerequisites:
#   - gnoland1 must be fully halted at the governance-approved halt height
#   - tx-archive (with genesis-assemble subcommand) must be installed
#     See: https://github.com/gnolang/tx-archive
#   - The new gnoland binary (>= gnoland-1) must be in PATH
#
# Dependencies (PRs that must be merged in the new binary):
#   - #5334: halt_height config field (MERGED)
#   - #5293: namereg GovDAO whitelist (MERGED)
#   - #5375: new govdao-scripts (MERGED)
#   - #5368: GovDAO-based halt height via r/sys/params (awaiting merge)
#   - #5411: genesis replay with OriginalChainID (awaiting merge)
#   - #5390: GnoTxMetadata block_height + chain_id extension (awaiting merge)
set -eo pipefail

# =============================================================================
# TODO: IMPLEMENT THIS MIGRATION SCRIPT
#
# This is the critical piece that makes the gnoland1 → gnoland-1 hard fork
# possible. Until it is written AND dry-run on test12, the hard fork CANNOT happen.
#
# The decisions below have been made (as of 2026-04-09):
#   ✅ Scenario A chosen: genesis tx-replay with InitialHeight preservation
#   ✅ Scenario B (height reset) deprioritized
#   ✅ Chain ID: gnoland1 → gnoland-1 (one-time rename, agreed)
#   ✅ Height + timestamp preservation required (Jae's correctness requirement)
#   ✅ #5334, #5293, #5375 merged
#   ⏳ #5368 (GovDAO halt) awaiting reviews
#   ⏳ #5411 (genesis replay), #5390 (tx metadata) awaiting merge
#   ⏳ #5377 (--migrate flag) may be superseded by #5411 approach
#
# Implementation steps:
#
#   1. HALT VERIFICATION
#      Check the data dir to confirm gnoland1 stopped at the expected height.
#      Read the committed block height from the blockstore and compare against
#      the halt_height that was voted on via GovDAO (#5368).
#
#   2. TX EXPORT
#      Run tx-archive backup against the halted gnoland1 data dir to produce
#      a JSONL file with all successful txs, each including:
#        - timestamp: Unix seconds of the block that included the tx
#        - block_height: block number the tx ran at (#5390)
#        - chain_id: "gnoland1" (#5390)
#      Command (once tx-archive supports --data-dir mode):
#        tx-archive backup \
#          --data-dir "$DATA_DIR" \
#          --output txs.jsonl
#
#   3. GENESIS ASSEMBLY
#      Run tx-archive genesis-assemble to produce genesis.json:
#        tx-archive genesis-assemble \
#          --input txs.jsonl \
#          --chain-id gnoland-1 \
#          --original-chain-id gnoland1 \
#          --initial-height $((HALT_HEIGHT + 1)) \
#          --output genesis.json
#      The genesis will have:
#        chain_id: "gnoland-1"
#        initial_height: halt_height + 1  (Jae's InitialHeight tm2 port required)
#        app_state.txs: full tx array with metadata
#        app_state.original_chain_id: "gnoland1" (for sig verification)
#
#   4. VERIFICATION
#      All validators independently run this script and compare:
#        sha256sum genesis.json
#      Hashes MUST match before anyone restarts.
#      Optionally run gnoupgrade statediff for before/after comparison.
#
#   5. RESTART COORDINATION
#      Validators restart with the new binary and the new genesis.json.
#      The new binary must be >= chain/gnoland-1 tag.
#
# BLOCKERS (must be resolved before implementing):
#   [ ] tx-archive genesis-assemble command (companion to #5411)
#   [ ] tx-archive --data-dir mode for offline export from block store
#   [ ] Jae's tm2 GenesisDoc.InitialHeight field (hard blocker for #5411)
#   [ ] test12 dry-run: validate full halt → export → genesis-assemble → restart
#
# RELATED:
#   Issue #5374: chain upgrade strategy meta-issue
#   PR #5411: genesis replay mechanism (main hardfork PR)
#   PR #5390: GnoTxMetadata block_height + chain_id extension
#   PR #5368: GovDAO-based halt height via r/sys/params
#   PR #5377: in-place block-replay --migrate flag (may complement #5411)
#   PR #5369: gnoupgrade toolkit (replay/statediff/healthcheck)
# =============================================================================

echo "ERROR: migrate-from-gnoland1.sh is not yet implemented."
echo ""
echo "Blockers:"
echo "  - tx-archive genesis-assemble command (companion to gnolang/gno#5411)"
echo "  - Jae's tm2 GenesisDoc.InitialHeight field"
echo "  - test12 dry-run of the full halt → export → genesis-assemble → restart flow"
echo ""
echo "Track progress at: https://github.com/gnolang/gno/issues/5374"
exit 1
