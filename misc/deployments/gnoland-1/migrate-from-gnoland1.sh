#!/usr/bin/env bash
# migrate-from-gnoland1.sh — hard fork gnoland1 → gnoland-1.
#
# This script produces a new genesis.json for the gnoland-1 chain from the
# final committed state of gnoland1 after it halts at the governance-approved
# halt height.
#
# Operators run this script once on their local node after gnoland1 has halted,
# then restart with the generated genesis and the new binary (>= gnoland-1).
#
# Usage:
#   ./migrate-from-gnoland1.sh --data-dir <gnoland1-data-dir>
#
# The script writes genesis.json (gnoland-1) to the current directory.
set -eo pipefail

# =============================================================================
# TODO: WRITE THIS MIGRATION SCRIPT
#
# This is the critical piece that makes the gnoland1 → gnoland-1 hard fork
# possible. Until it is written, the hard fork CANNOT happen.
#
# The script must:
#
#   1. HALT VERIFICATION
#      - Confirm that gnoland1 has halted at the expected halt height
#        (set via GovDAO proposal using params.NewSetHaltRequest).
#      - Read the last committed block height and app_hash from the data dir.
#      - Compare against the expected halt_height stored in node:p:halt_height.
#
#   2. STATE EXPORT
#      - Export the full committed state from the gnoland1 data directory:
#          balances, realm state, packages, validator set, governance state.
#      - Produce a canonical sorted dump suitable for diffing (see Tool 3 in CI strategy).
#      - Verify the dump's hash matches the expected halt app_hash
#        (agreed by validators off-chain after the halt).
#
#   3. STATE MIGRATION TRANSFORMS
#      Apply the following changes to the exported state before producing the
#      new genesis. Each transform must be reversible and auditable:
#
#      3a. Chain ID rename
#          - Change chain_id from "gnoland1" to "gnoland-1" in all relevant state.
#
#      3b. r/sys/params upgrade (gnolang/gno#5368)
#          - The halt_height and halt_min_version params were added by the new
#            binary and are already in state. Verify they are present.
#          - No realm code migration needed — the new binary ships the updated
#            r/sys/params realm code.
#
#      3c. r/gnops/valopers fix (gnolang/gno#5373 — valoper price → 0)
#          - Verify that the min_fee param is 0 (should have been set via GovDAO
#            before the halt). If not, apply the change here.
#
#      3d. Namereg govdao whitelist (gnolang/gno#5293)
#          - Include the namereg PR changes if merged before the halt.
#          - If not merged, apply the namereg state patch here.
#          TODO: confirm which PRs are bundled in the hard fork binary.
#
#      3e. Gas parameter updates (gnolang/gno#5291, #5289, #5274)
#          - The new binary carries updated gas params. No state migration
#            needed for the params themselves, but:
#          - Any previously-valid transactions that are now invalid under the
#            new gas schedule must be identified and communicated.
#          TODO: enumerate which (if any) historical txs become invalid.
#
#   4. GENESIS ASSEMBLY
#      - Produce a new genesis.json with:
#          chain_id: "gnoland-1"
#          genesis_time: <halt block timestamp>
#          initial_height: <halt block height + 1>  OR  0 (TBD — see open question below)
#          app_state: <migrated state from step 3>
#          validators: <same validator set as gnoland1 at halt>
#      TODO: decide on initial_height (preserve vs reset — Scenario A vs B).
#      See tasks/chain-upgrade/CLAUDE.md for the Jae / Thomas discussion.
#
#   5. VERIFICATION
#      - All validators independently run this script and compare their
#        genesis.json SHA-256. Hashes must match before anyone restarts.
#      - Run the integration test suite (Tool 2) against the new genesis to
#        verify realm state is consistent.
#      - Optionally: run the state diff tool (Tool 3) to confirm only expected
#        changes were applied.
#
#   6. RESTART COORDINATION
#      - Validators restart with the new binary (>= chain/gnoland-1) and the
#        new genesis.json.
#      - The new binary enforces node:p:halt_min_version — old binaries will
#        refuse to start, preventing accidental resumption of gnoland1.
#      - First block on gnoland-1 must be signed by >2/3 of the validator set.
#
# OPEN QUESTIONS (resolve before implementing):
#   [ ] Scenario A vs B: preserve height from gnoland1 or reset to 0?
#       Jae's concern: in-place migration is riskier if the script is faulty
#       (divergence between nodes). A new genesis (height=0) is safer but
#       worse UX. See CLAUDE.md for full discussion.
#   [ ] Timestamp handling: if preserving height, how do we handle per-tx
#       timestamps? Jae requires exact preservation. This needs a special dump
#       format that captures block timestamps alongside each tx.
#   [ ] Which PRs are bundled in the hard fork binary? Confirm with the team:
#         - #5334 (halt-height CLI flag) — merged?
#         - #5368 (GovDAO halt + min_version) — merged?
#         - #5293 (namereg govdao whitelist) — merged?
#         - #5291, #5289, #5274 (gas params) — merged?
#   [ ] Who runs this script and when? Each validator independently, or
#       one validator produces the genesis and others verify?
#
# RELATED:
#   See also: tasks/chain-upgrade/CLAUDE.md (pyxis repo) for full context.
#   PRs: gnolang/gno#5334, #5368, #5373, #5293
# =============================================================================

echo "ERROR: migrate-from-gnoland1.sh is not yet implemented."
echo ""
echo "This migration script is the critical missing piece for the gnoland1 → gnoland-1"
echo "hard fork. See the TODO block at the top of this file for what needs to be written."
echo ""
echo "Track progress at: https://github.com/gnolang/gno/issues/TODO"
exit 1
