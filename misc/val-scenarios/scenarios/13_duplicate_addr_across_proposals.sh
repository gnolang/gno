#!/usr/bin/env bash
set -euo pipefail

SCENARIO_CI=true

# Scenario 13: duplicate validator address across two separate proposals in the
# same block.
#
# A single MsgRun executes two proposals that individually are valid:
#   Proposal 1: remove val1 (VotingPower: 0)
#   Proposal 2: re-add val1 (VotingPower: 5)
#
# The EndBlocker deduplicates per-block changes (second proposal overwrites the
# first), so val1 ends up with VotingPower=5 and the chain keeps advancing.

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

cat >"${script_dir}/two_proposals.gno" <<GNOEOF
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

cat >"${script_dir}/assert_validator_added.gno" <<GNOEOF
package main

import valr "gno.land/r/sys/validators/v2"

func main() {
	addr := address("${TARGET_ADDR}")
	if !valr.IsValidator(addr) {
		panic("val4 validator was not added")
	}

	val := valr.GetValidator(addr)
	if val.PubKey != "${TARGET_PUBKEY}" {
		panic("val4 validator pubkey mismatch")
	}
	if val.VotingPower != ${TARGET_POWER} {
		panic("val4 validator voting power mismatch")
	}
}
GNOEOF

log "estimating gas for the two-proposal script"
set +e
run_gas="$(estimate_run_gas val1 "${script_dir}/two_proposals.gno" 200000000)"
estimate_status=$?
set -e
if [ "$estimate_status" -ne 0 ]; then
  run_gas=200000000
  log "gas estimation failed; using fallback gas=${run_gas}"
else
  log "gas estimate: ${run_gas}"
fi

log "submitting two-proposal script"
run_script val1 "${script_dir}/two_proposals.gno" "$run_gas"

run_script val1 "${script_dir}/assert_validator_added.gno" 50000000 only >/dev/null
assert_chain_advances val1 120 5

print_cluster_status
