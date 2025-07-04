package main

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	users "gno.land/r/gnoland/users/v1"
	"gno.land/r/gov/dao"
	daov3init "gno.land/r/gov/dao/v3/init"
	susers "gno.land/r/sys/users"
)

// Test giving an arbi name via GovDAO

var c std.Address = std.OriginCaller()

func init() {
	daov3init.InitWithUsers(c)

	alice := testutils.TestAddress("alice")

	// Register alice
	testing.SetOriginCaller(alice)
	testing.SetRealm(std.NewUserRealm(alice))
	users.Register(cross, "alice123")

	// Prop to delete user
	testing.SetOriginCaller(c)
	testing.SetRealm(std.NewUserRealm(c))
	pr := users.ProposeArbitraryName("alice", "we like alice so we want to give her an arbitrary name", alice)
	dao.MustCreateProposal(cross, pr)
}

func main() {
	testing.SetOriginCaller(c)

	println("--")
	println(dao.Render(""))
	println("--")
	println(dao.Render("0"))
	println("--")

	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(0),
	})

	println("--")
	println(dao.Render("0"))
	println("--")

	dao.ExecuteProposal(cross, dao.ProposalID(0))

	println("--")
	println(dao.Render("0"))

	println("--")

	data, _ := susers.ResolveName("alice")
	println(data.Addr())
}

// Output:
// --
// # Active Proposals:
// ## Proposal with id: 0
// ### Title: User Registry V1: Give `alice` to `g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh`
//
// ### Proposed by: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// we like alice so we want to give her an arbitrary name
//
//
// ### Proposal Status:
//
// - **Proposal open for votes**
// - Allowed tiers to vote: T1 T2 T3
// - YES PERCENT: 0%
// - NO PERCENT: 0%
// - [Go to votes list](/r/gov/dao:0/votes).
//
//
// --
// ## Proposal with id: 0
// ### Title: User Registry V1: Give `alice` to `g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh`
//
// ### Proposed by: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// we like alice so we want to give her an arbitrary name
//
//
// ### Proposal Status:
//
// - **Proposal open for votes**
// - Allowed tiers to vote: T1 T2 T3
// - YES PERCENT: 0%
// - NO PERCENT: 0%
// - [Go to votes list](/r/gov/dao:0/votes).
//
//
// --
// --
// ## Proposal with id: 0
// ### Title: User Registry V1: Give `alice` to `g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh`
//
// ### Proposed by: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// we like alice so we want to give her an arbitrary name
//
//
// ### Proposal Status:
//
// - **Proposal open for votes**
// - Allowed tiers to vote: T1 T2 T3
// - YES PERCENT: 100%
// - NO PERCENT: 0%
// - [Go to votes list](/r/gov/dao:0/votes).
//
//
// --
// --
// ## Proposal with id: 0
// ### Title: User Registry V1: Give `alice` to `g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh`
//
// ### Proposed by: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// we like alice so we want to give her an arbitrary name
//
//
// ### Proposal Status:
//
// - **PROPOSAL HAS BEEN ACCEPTED**
// - Allowed tiers to vote: T1 T2 T3
// - YES PERCENT: 100%
// - NO PERCENT: 0%
// - [Go to votes list](/r/gov/dao:0/votes).
//
//
// --
// g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh
