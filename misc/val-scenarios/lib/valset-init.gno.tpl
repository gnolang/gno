// valset-init.gno is executed via MsgRun during genesis to register
// the initial validator set in the on-chain PoA.
//
// Steps:
//  1. Add the genesis deployer as sole T1 govDAO member (100% supermajority).
//  2. Register all test validators via governance proposal.
//  3. Remove the deployer — leaves the DAO empty for tests.
package main

import (
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
	valr "gno.land/r/sys/validators/v2"
)

// genesisDeployerAddr is derived from the deployer mnemonic used by generate_genesis.
const genesisDeployerAddr = address("g1edq4dugw0sgat4zxcw9xardvuydqf6cgleuc8p")

func main(cur realm) {
	ms := memberstore.Get(0, cur)

	// 1. Make deployer sole T1 member so proposals pass immediately.
	must(ms.SetMember(memberstore.T1, genesisDeployerAddr, memberstore.NewMember(0)))

	// 2. Register the test validator set.
	r := valr.NewPropRequest(
		cross(cur),
		func() []validators.Validator {
			return []validators.Validator{
				// GEN:VALSET
			}
		},
		"Add initial test validator set",
		"",
	)
	pid := dao.MustCreateProposal(cross(cur), r)
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))
	dao.ExecuteProposal(cross(cur), pid)

	// 3. Remove deployer — leaves an empty, unlocked DAO for test use.
	ms.RemoveMember(genesisDeployerAddr)
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
