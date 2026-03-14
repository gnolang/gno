// govdao_prop1.gno is executed via MsgRun during genesis to bootstrap the
// gnoland1 governance state.
//
// At this point the loader has initialised the DAO and tier structure but
// has NOT added any members and has NOT set AllowedDAOs (the empty-list
// corner case in proxy.gno:InAllowedDAOs lets any caller through).
//
// NOTE: Temporarily, manfred (moul) is the sole govDAO T1 member. He will
// craft proposals to add validators and other members, executing them alone
// until the DAO is large enough to enable real multi-party voting.
//
// Steps:
//  1. Add a temporary deployer as T1 member (100% supermajority).
//  2. Register the initial validator set via governance proposal.
//  3. Set chain params (unrestricted addrs, restricted denoms) via proposals.
//  4. Enable namespace enforcement (r/sys/names).
//  5. Add the real govDAO T1 member (manfred) directly to the memberstore.
//  6. Remove the temporary deployer from the memberstore.
//  7. Lock down memberstore access by setting AllowedDAOs.
package main

import (
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
	"gno.land/r/sys/names"
	"gno.land/r/sys/params"
	valr "gno.land/r/sys/validators/v2"
)

// genesisDeployerAddr is the address derived from the genesis deployer
// mnemonic.  See loader.gno for the full mnemonic.
const genesisDeployerAddr = address("g1edq4dugw0sgat4zxcw9xardvuydqf6cgleuc8p")

func main() {
	ms := memberstore.Get()

	// ---- 1. Add temporary deployer as sole T1 member ----
	// This gives the deployer 100% supermajority so proposals pass.
	must(ms.SetMember(memberstore.T1, genesisDeployerAddr, &memberstore.Member{InvitationPoints: 0}))

	// ---- 2. Register the initial validator set via governance proposal ----
	// 7 validators — BFT >2/3 threshold (floor(2n/3)+1) means 5 nodes must be up for consensus.
	// Additional validators can be added post-genesis via govDAO proposals
	// (see govdao-scripts/add-validator.sh).
	govExec(valr.NewPropRequest(
		func() []validators.Validator {
			return []validators.Validator{
				{
					Address:     address("g1uhv7wr7nku89se3t7v8fpquc7n5sf8rfkywxpc"),
					PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqavtgten8l8k4f72j8klpu4l7tk2qw4kl8394krsaysmz2q0765ynvjag0",
					VotingPower: 1,
				},
				{
					Address:     address("g10jdd8vlgydfypynrk23ul90jnsg5twrtvmcmh4"),
					PubKey:      "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqve8jffvhy97sfc5gyvag09h8g9g3d9e4cta7s7m6vcmzug84kjywg7fn2y",
					VotingPower: 1,
				},
				{
					Address:     address("g1eueypc9w524ctda3y0kwd4jruw5p4zqpjna0jq"),
					PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqsdatvn76ru9pck7zrt7zc33y8p2kweaujx82utgmtwljypuz9p8uctk3m",
					VotingPower: 1,
				},
				{
					Address:     address("g1kn7p0wqumvqlcqzhkwnavkhf0z4qnr73ltwsae"),
					PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpufsm93d5fmzxrug76esaxsdsmw0guy9e6geypw7ekz92sl3mte492q62e",
					VotingPower: 1,
				},
			}
		},
		"Add initial validator set",
		"Bootstrap the gnoland1 initial validator set",
	))

	// ---- 3. Set chain parameters via governance proposals ----

	// Unrestricted addresses (can transfer ugnot even when bank is locked).
	govExec(params.ProposeAddUnrestrictedAcctsRequest(
		address("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"), // GovDAO T1 multisig
		address("g148583t5x66zs6p90ehad6l4qefeyaf54s69wql"), // faucet
	))

	// Lock ugnot transfers (restricted denom).
	govExec(params.ProposeLockTransferRequest())

	// ---- 4. Enable namespace enforcement (r/sys/names) ----
	// This drops ownership and enables the namespace permission check.
	// Must be called during genesis; after this, package paths require registration.
	names.Enable(cross)

	// ---- 5. Add govDAO T1 member ----
	// Temporarily manfred (moul) is the sole govDAO member. He will manually
	// craft and execute proposals to onboard validators and additional members
	// until the DAO is large enough for real multi-party governance.
	must(ms.SetMember(memberstore.T1, address("g1manfred47kzduec920z88wfr64ylksmdcedlf5"), &memberstore.Member{InvitationPoints: 3})) // moul

	// ---- 6. Remove the genesis deployer ----
	ms.RemoveMember(genesisDeployerAddr)

	// ---- 7. Lock down memberstore access ----
	// From this point on, only gno.land/r/gov/dao/v3/impl can call
	// memberstore.Get() — matching the production AllowedDAOs configuration.
	dao.UpdateImpl(cross, dao.UpdateRequest{
		AllowedDAOs: []string{"gno.land/r/gov/dao/v3/impl"},
	})
}

// govExec creates a proposal, votes YES, and executes it immediately.
// Works because the deployer is the sole T1 member (100% supermajority).
func govExec(r dao.ProposalRequest) {
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
