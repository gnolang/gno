#!/usr/bin/env bash
set -euo pipefail

SCENARIO_CI=true

# Scenario 17: add and remove a validator through GovDAO proposals.
#
# The scenario starts a 3-validator genesis set plus a synced val4 node that is
# not part of the initial validator set. It adds val4 through a GovDAO proposal,
# verifies the on-chain validator set, then removes val4 through a second GovDAO
# proposal.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/lib/scenario.sh"

scenario_init "scenario-17"
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

cat > "${script_dir}/add_validator.gno" << GNOEOF
package main

import (
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
	valr "gno.land/r/sys/validators/v2"
)

const txAddr = address("${TX_ADDRESS}")

func main() {
	// The scenario genesis leaves allowedDAOs empty, so the local transaction
	// key can bootstrap itself as a GovDAO T1 member for this proposal.
	must(memberstore.Get().SetMember(memberstore.T1, txAddr, &memberstore.Member{InvitationPoints: 0}))

	r := valr.NewPropRequest(
		func() []validators.Validator {
			return []validators.Validator{
				{
					Address:     address("${VAL4_ADDR}"),
					PubKey:      "${VAL4_PUBKEY}",
					VotingPower: ${VAL4_POWER},
				},
			}
		},
		"Add validator val4",
		"Add val4 (${VAL4_ADDR}) with power ${VAL4_POWER}",
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

cat > "${script_dir}/assert_validator_added.gno" << GNOEOF
package main

import valr "gno.land/r/sys/validators/v2"

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

cat > "${script_dir}/rm_validator.gno" << GNOEOF
package main

import (
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	valr "gno.land/r/sys/validators/v2"
)

func main() {
	r := valr.NewPropRequest(
		func() []validators.Validator {
			return []validators.Validator{
				{
					Address:     address("${VAL4_ADDR}"),
					VotingPower: 0,
				},
			}
		},
		"Remove validator val4",
		"Remove val4 (${VAL4_ADDR}) from the validator set",
	)
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GNOEOF

cat > "${script_dir}/assert_validator_removed.gno" << GNOEOF
package main

import valr "gno.land/r/sys/validators/v2"

func main() {
	if valr.IsValidator(address("${VAL4_ADDR}")) {
		panic("val4 validator was not removed")
	}
}
GNOEOF

log "estimating gas for add-validator proposal"
set +e
add_gas="$(estimate_run_gas val1 "${script_dir}/add_validator.gno" 50000000)"
estimate_status=$?
set -e
if [ "$estimate_status" -ne 0 ]; then
  add_gas=50000000
  log "gas estimation failed; using fallback gas=${add_gas}"
else
  log "gas estimate: ${add_gas}"
fi

log "submitting add-validator GovDAO proposal"
run_script val1 "${script_dir}/add_validator.gno" "$add_gas"
assert_chain_advances val1 120 3

log "verifying val4 is in the on-chain validator set"
run_script val1 "${script_dir}/assert_validator_added.gno" 50000000 only >/dev/null
assert_chain_advances val4 120 2

# Keep the add and remove changes in different blocks so this scenario tests the
# normal operational flow instead of the duplicate-change edge cases covered by
# scenarios 12 and 13.
wait_for_blocks val1 2 120

log "estimating gas for remove-validator proposal"
set +e
rm_gas="$(estimate_run_gas val1 "${script_dir}/rm_validator.gno" 50000000)"
estimate_status=$?
set -e
if [ "$estimate_status" -ne 0 ]; then
  rm_gas=50000000
  log "gas estimation failed; using fallback gas=${rm_gas}"
else
  log "gas estimate: ${rm_gas}"
fi

log "submitting remove-validator GovDAO proposal"
run_script val1 "${script_dir}/rm_validator.gno" "$rm_gas"
assert_chain_advances val1 120 3

log "verifying val4 is no longer in the on-chain validator set"
run_script val1 "${script_dir}/assert_validator_removed.gno" 50000000 only >/dev/null

print_cluster_status
