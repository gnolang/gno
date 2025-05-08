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
	cross(users.Register)("alice123")

	// Prop to delete user
	testing.SetOriginCaller(c)
	testing.SetRealm(std.NewUserRealm(c))
	pr := users.ProposeDeleteUser(alice, "delete user test")
	cross(dao.MustCreateProposal)(pr)
}

func main() {
	testing.SetOriginCaller(c)

	println("--")
	println(dao.Render(""))
	println("--")
	println(dao.Render("0"))
	println("--")

	cross(dao.MustVoteOnProposal)(dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(0),
	})

	println("--")
	println(dao.Render("0"))
	println("--")

	cross(dao.ExecuteProposal)(dao.ProposalID(0))

	println("--")
	println(dao.Render("0"))

	data, _ := susers.ResolveName("alice123")
	if data == nil {
		println("Successfully deleted alice")
	}
}

// Output:
// --
// # Active Proposals:
// ## Proposal with id: 0
// ### Title: Propose deleting a name using users/v1 realm
//
// ### Proposed by: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// delete user test
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
// ### Title: Propose deleting a name using users/v1 realm
//
// ### Proposed by: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// delete user test
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
// ### Title: Propose deleting a name using users/v1 realm
//
// ### Proposed by: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// delete user test
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
// ### Title: Propose deleting a name using users/v1 realm
//
// ### Proposed by: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// delete user test
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
// Successfully deleted alice
