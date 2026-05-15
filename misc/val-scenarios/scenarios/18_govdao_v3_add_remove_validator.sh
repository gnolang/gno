#!/usr/bin/env bash
set -euo pipefail

SCENARIO_CI=true

# Scenario 18: add and remove a validator through r/sys/validators/v3.
#
# This is the v3 variant of scenario 17. It starts a 3-validator genesis set
# plus a synced val4 node that is not part of the initial validator set. It
# registers val4 as a valoper in r/gnops/valopers (TX_ADDRESS as operator,
# val4's consensus pubkey as the signing key), then adds val4 through a GovDAO
# proposal using r/sys/validators/v3.NewValidatorProposalRequest, verifies the
# v3 on-chain validator set, then removes val4 through a second v3 proposal.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/lib/scenario.sh"

scenario_init "scenario-18"
trap scenario_finish EXIT

gen_validator val1
gen_validator val2
gen_validator val3
gen_validator val4 --not-in-genesis

prepare_network

VAL4_ADDR="${NODE_ADDRESS[val4]}"
VAL4_PUBKEY="${NODE_PUBKEY[val4]}"
VAL4_POWER="${NODE_POWER[val4]}"
log "val4 candidate validator: ${VAL4_ADDR} power=${VAL4_POWER}"

start_all_nodes
assert_chain_advances val1 120 5
assert_chain_advances val4 120 2

script_dir="${SCENARIO_DIR}/scripts"
mkdir -p "$script_dir"

cat >"${script_dir}/add_validator_v3.gno" <<GNOEOF
package main

import (
	"gno.land/r/gnops/valopers"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
	valr "gno.land/r/sys/validators/v3"
)

const txAddr = address("${TX_ADDRESS}")

func main() {
	// Register val4's consensus key with TX_ADDRESS as operator so v3's
	// valoperCache is populated before the proposal.
	valopers.Register(cross, "val4", "val4 test validator", "cloud", txAddr, "${VAL4_PUBKEY}")

	// The scenario genesis leaves allowedDAOs empty, so the local transaction
	// key can bootstrap itself as a GovDAO T1 member for this proposal.
	must(memberstore.Get().SetMember(memberstore.T1, txAddr, &memberstore.Member{InvitationPoints: 0}))

	r := valr.NewValidatorProposalRequest(
		[]valr.ValoperChange{
			{
				OperatorAddress: txAddr,
				Power:           ${VAL4_POWER},
			},
		},
		"Add validator val4 with validators v3",
		"Add val4 (${VAL4_ADDR}) with power ${VAL4_POWER} through r/sys/validators/v3",
	)
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
GNOEOF

cat >"${script_dir}/assert_validator_added_v3.gno" <<GNOEOF
package main

import valr "gno.land/r/sys/validators/v3"

func main() {
	addr := address("${VAL4_ADDR}")
	if !valr.IsValidator(addr) {
		panic("val4 validator was not added")
	}

	val := valr.GetValidator(addr)
	if val.PubKey != "${VAL4_PUBKEY}" {
		panic("val4 validator pubkey mismatch")
	}
	if val.VotingPower != ${VAL4_POWER} {
		panic("val4 validator voting power mismatch")
	}
}
GNOEOF

cat >"${script_dir}/rm_validator_v3.gno" <<GNOEOF
package main

import (
	"gno.land/r/gov/dao"
	valr "gno.land/r/sys/validators/v3"
)

func main() {
	r := valr.NewValidatorProposalRequest(
		[]valr.ValoperChange{
			{
				OperatorAddress: address("${TX_ADDRESS}"),
				Power:           0,
			},
		},
		"Remove validator val4 with validators v3",
		"Remove val4 (${VAL4_ADDR}) from the validator set through r/sys/validators/v3",
	)
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GNOEOF

cat >"${script_dir}/assert_validator_removed_v3.gno" <<GNOEOF
package main

import valr "gno.land/r/sys/validators/v3"

func main() {
	if valr.IsValidator(address("${VAL4_ADDR}")) {
		panic("val4 validator was not removed")
	}
}
GNOEOF

log "estimating gas for v3 add-validator proposal"
set +e
add_gas="$(estimate_run_gas val1 "${script_dir}/add_validator_v3.gno" 200000000)"
estimate_status=$?
set -e
if [ "$estimate_status" -ne 0 ]; then
  add_gas=200000000
  log "gas estimation failed; using fallback gas=${add_gas}"
else
  log "gas estimate: ${add_gas}"
fi

log "submitting v3 add-validator GovDAO proposal"
run_script val1 "${script_dir}/add_validator_v3.gno" "$add_gas"
assert_chain_advances val1 120 3

log "verifying val4 is in the v3 on-chain validator set"
run_script val1 "${script_dir}/assert_validator_added_v3.gno" 50000000 only >/dev/null
assert_chain_advances val4 120 2

# Keep the add and remove changes in different blocks so this scenario tests the
# normal operational flow instead of duplicate-change edge cases.
wait_for_blocks val1 2 120

log "estimating gas for v3 remove-validator proposal"
set +e
rm_gas="$(estimate_run_gas val1 "${script_dir}/rm_validator_v3.gno" 200000000)"
estimate_status=$?
set -e
if [ "$estimate_status" -ne 0 ]; then
  rm_gas=200000000
  log "gas estimation failed; using fallback gas=${rm_gas}"
else
  log "gas estimate: ${rm_gas}"
fi

log "submitting v3 remove-validator GovDAO proposal"
run_script val1 "${script_dir}/rm_validator_v3.gno" "$rm_gas"
assert_chain_advances val1 120 3

log "verifying val4 is no longer in the v3 on-chain validator set"
run_script val1 "${script_dir}/assert_validator_removed_v3.gno" 50000000 only >/dev/null

print_cluster_status
