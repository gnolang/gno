#!/usr/bin/env bash
set -euo pipefail

# Disabled in CI: outdated, needs to be updated.
SCENARIO_CI=false

# Scenario 13: duplicate validator address across two separate proposals in the
# same block.
#
# A single MsgRun executes two proposals that individually are valid:
#   Proposal 1: remove val1 (VotingPower: 0)
#   Proposal 2: re-add val1 (VotingPower: 5)
#
# Each proposal passes its own validation, but saveChange in
# r/sys/validators/v2/validators.gno blindly appends every change to the
# per-block slice. The block-level aggregate therefore contains two entries for
# the same address [{val1, 0}, {val1, 5}].
#
# At EndBlocker, GetChanges returns that aggregate. tm2 processChanges detects
# the duplicate, ApplyBlock fails, and the node is killed (osm.Kill).
#
# val1 should end up with VotingPower=5 and the chain should keep advancing.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/lib/scenario.sh"

scenario_init "scenario-13"
trap scenario_finish EXIT

gen_validator val1

prepare_network
start_all_nodes
assert_chain_advances val1 120 5

TARGET="val1"
TARGET_ADDR="${NODE_ADDRESS[$TARGET]}"
TARGET_PUBKEY="${NODE_PUBKEY[$TARGET]}"
TARGET_POWER=5

script_dir="${SCENARIO_DIR}/scripts"
mkdir -p "$script_dir"

cat > "${script_dir}/two_proposals.gno" << GNOEOF
package main

import (
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
	valr "gno.land/r/sys/validators/v2"
)

const txAddr = address("${TX_ADDRESS}")

func main() {
	must(memberstore.Get().SetMember(memberstore.T1, txAddr, &memberstore.Member{InvitationPoints: 0}))

	// Proposal 1: remove val1 — individually valid.
	r1 := valr.NewPropRequest(
		func() []validators.Validator {
			return []validators.Validator{
				{
					Address:     address("${TARGET_ADDR}"),
					VotingPower: 0,
				},
			}
		},
		"Remove validator ${TARGET_ADDR}",
		"",
	)
	pid1 := dao.MustCreateProposal(cross, r1)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid1})
	dao.ExecuteProposal(cross, pid1)

	// Proposal 2: re-add val1 with new power — individually valid.
	// Together with proposal 1, the block-level change slice now contains
	// two entries for the same address, which crashes the node in EndBlocker.
	r2 := valr.NewPropRequest(
		func() []validators.Validator {
			return []validators.Validator{
				{
					Address:     address("${TARGET_ADDR}"),
					PubKey:      "${TARGET_PUBKEY}",
					VotingPower: ${TARGET_POWER},
				},
			}
		},
		"Re-add validator ${TARGET_ADDR}",
		"",
	)
	pid2 := dao.MustCreateProposal(cross, r2)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid2})
	dao.ExecuteProposal(cross, pid2)
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
GNOEOF

log "estimating gas for the two-proposal script"
set +e
run_gas="$(estimate_run_gas val1 "${script_dir}/two_proposals.gno" 50000000)"
estimate_status=$?
set -e
if [ "$estimate_status" -ne 0 ]; then
  run_gas=50000000
  log "gas estimation failed; using fallback gas=${run_gas}"
else
  log "gas estimate: ${run_gas}"
fi

log "submitting two-proposal script"
run_script val1 "${script_dir}/two_proposals.gno" "$run_gas"

# BUG: once saveChange deduplicates per-block changes, the second proposal
# should overwrite the first and val1 should end up with VotingPower=5, with
# the chain advancing normally. On unpatched master, both saveChange calls
# succeed, GetChanges returns two entries for the same address, processChanges
# in tm2 detects the duplicate, ApplyBlock fails, and the node is killed via
# osm.Kill. Assert the known-buggy behaviour so CI stays green until the fix
# lands. When fixed, replace the line below with:
#   assert_chain_advances val1 120 5
assert_chain_halted val1 120

print_cluster_status
