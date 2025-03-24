package main

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	users "gno.land/r/gnoland/users/v1"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/init"
	susers "gno.land/r/sys/users"
)

// Test updating a name via GovDAO
var c std.Address = std.OriginCaller()

func init() {
	init.InitWithUsers(c)

	alice := testutils.TestAddress("alice")

	// Register alice
	testing.SetOriginCaller(alice)
	testing.SetRealm(std.NewUserRealm(alice))
	users.Register("alice123")

	// Prop to change name
	testing.SetOriginCaller(c)
	testing.SetRealm(std.NewUserRealm(c))
	testing.SetOriginPkgAddress(std.Address("g1p84dvfh4wrplyxx4zsmy77a8rxcnjup2j7zv5r")) // gno.land/r/gov/dao
	pr := users.ProposeNewName(alice, "alice_new123")
	dao.MustCreateProposal(pr)
}

func main() {
	testing.SetOriginCaller(c)
	testing.SetOriginPkgAddress(std.Address("g1p84dvfh4wrplyxx4zsmy77a8rxcnjup2j7zv5r")) // gno.land/r/gov/dao

	println("--")
	println(dao.Render(""))
	println("--")
	println(dao.Render("0"))
	println("--")

	dao.MustVoteOnProposal(dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(0),
	})
	println("--")
	println(dao.Render("0"))
	println("--")
	dao.ExecuteProposal(dao.ProposalID(0))
	println("--")
	println(dao.Render("0"))

	data, _ := susers.ResolveName("alice_new123")
	println(data.Addr())
}

// Output:
// --
// # Active Proposals:
// ## Proposal with id: 0
// ### Title: Propose a new name using users/v1 realm
//
// ### Proposed by: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
//
//
// This proposal contains the following metadata:
//
// No metadata available.
//
// ### Proposal Status:
//
// - **Proposal open for votes**
// - Allowed tiers to vote: T1 T2 T3
// - YES PERCENT: 0%
// - NO PERCENT: 0%
// - [Go to votes list](:0/votes).
//
//
// --
// ## Proposal with id: 0
// ### Title: Propose a new name using users/v1 realm
//
// ### Proposed by: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
//
//
// This proposal contains the following metadata:
//
// No metadata available.
//
// ### Proposal Status:
//
// - **Proposal open for votes**
// - Allowed tiers to vote: T1 T2 T3
// - YES PERCENT: 0%
// - NO PERCENT: 0%
// - [Go to votes list](:0/votes).
//
//
// --
// --
// ## Proposal with id: 0
// ### Title: Propose a new name using users/v1 realm
//
// ### Proposed by: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
//
//
// This proposal contains the following metadata:
//
// No metadata available.
//
// ### Proposal Status:
//
// - **Proposal open for votes**
// - Allowed tiers to vote: T1 T2 T3
// - YES PERCENT: 100%
// - NO PERCENT: 0%
// - [Go to votes list](:0/votes).
//
//
// --
// --
// ## Proposal with id: 0
// ### Title: Propose a new name using users/v1 realm
//
// ### Proposed by: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
//
//
// This proposal contains the following metadata:
//
// No metadata available.
//
// ### Proposal Status:
//
// - **PROPOSAL HAS BEEN ACCEPTED**
// - Allowed tiers to vote: T1 T2 T3
// - YES PERCENT: 100%
// - NO PERCENT: 0%
// - [Go to votes list](:0/votes).
//
//
// g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh
