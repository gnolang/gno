// govdao_prop1.gno is executed via MsgRun during genesis to bootstrap the
// betanet governance state.
//
// At this point the loader has initialised the DAO and tier structure but
// has NOT added any members and has NOT set AllowedDAOs (the empty-list
// corner case in proxy.gno:InAllowedDAOs lets any caller through).
//
// Steps:
//  1. Add a temporary deployer as T1 member (100% supermajority).
//  2. Create a proposal to register the initial validator set, vote YES,
//     and execute it immediately.
//  3. Add the real govDAO T1 members directly to the memberstore.
//  4. Remove the temporary deployer from the memberstore.
//  5. Lock down memberstore access by setting AllowedDAOs.
package main

import (
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
	valr "gno.land/r/sys/validators/v2"
)

// genesisDeployerAddr is the address derived from the genesis deployer
// mnemonic.  See loader.gno for the full mnemonic.
const genesisDeployerAddr = address("g1edq4dugw0sgat4zxcw9xardvuydqf6cgleuc8p")

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	ms := memberstore.Get()

	// ---- 1. Add temporary deployer as sole T1 member ----
	// This gives the deployer 100% supermajority so proposals pass.
	must(ms.SetMember(memberstore.T1, genesisDeployerAddr, &memberstore.Member{InvitationPoints: 0}))

	// ---- 2. Register the initial validator set via governance proposal ----
	pid := dao.MustCreateProposal(cross, valr.NewPropRequest(
		func() []validators.Validator {
			return []validators.Validator{
				{
					Address:     address("g1e8j30772p4fvjd27eqq0e50vs74ekst50f5hz6"),
					PubKey:      "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqdw9mje8pwpd8p6pu3h0ke9mxp3s7l5vr6jtsffyd9wazpgxufwrssj367n",
					VotingPower: 1,
				},
				{
					Address:     address("g18k5nz67cg03lrdkwnf9qmdfzf0ptaq7lxefe4y"),
					PubKey:      "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0mxx3hhmy5hsl9kv8ne2ch46l8uk3kpnwc8kr900066ejz3zlhqcj6wjna",
					VotingPower: 1,
				},
			}
		},
		"Add initial validator set",
		"Bootstrap the betanet initial validator set",
	))
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)

	// ---- 3. Add govDAO T1 members ----
	must(ms.SetMember(memberstore.T1, address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"), &memberstore.Member{InvitationPoints: 3})) // Jae
	must(ms.SetMember(memberstore.T1, address("g1manfred47kzduec920z88wfr64ylksmdcedlf5"), &memberstore.Member{InvitationPoints: 3})) // moul
	must(ms.SetMember(memberstore.T1, address("g1m0rgan0rla00ygmdmp55f5m0unvsvknluyg2a4"), &memberstore.Member{InvitationPoints: 3})) // Morgan
	must(ms.SetMember(memberstore.T1, address("g1mx4pum9976th863jgry4sdjzfwu03qan5w2v9j"), &memberstore.Member{InvitationPoints: 3})) // Ray
	must(ms.SetMember(memberstore.T1, address("g12vx7dn3dqq89mz550zwunvg4qw6epq73d9csay"), &memberstore.Member{InvitationPoints: 3})) // Dongwon
	must(ms.SetMember(memberstore.T1, address("g127l4gkhk0emwsx5tmxe96sp86c05h8vg5tufzq"), &memberstore.Member{InvitationPoints: 3})) // Maxwell
	must(ms.SetMember(memberstore.T1, address("g1e6gxg5tvc55mwsn7t7dymmlasratv7mkv0rap2"), &memberstore.Member{InvitationPoints: 3})) // Milos

	// ---- 4. Remove the genesis deployer ----
	ms.RemoveMember(genesisDeployerAddr)

	// ---- 5. Lock down memberstore access ----
	// From this point on, only gno.land/r/gov/dao/v3/impl can call
	// memberstore.Get() — matching the production AllowedDAOs configuration.
	dao.UpdateImpl(cross, dao.UpdateRequest{
		AllowedDAOs: []string{"gno.land/r/gov/dao/v3/impl"},
	})
}
