package main

// SEND: 1000000ugnot

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/testutils"
	users "gno.land/r/gnoland/users/v1"
	"gno.land/r/gov/dao"
	daov3init "gno.land/r/gov/dao/v3/init"
	susers "gno.land/r/sys/users"
)

// Test updating a name via GovDAO
var c address = runtime.OriginCaller()

func init() {
	daov3init.InitWithUsers(c)

	alice := testutils.TestAddress("alice")

	// Register alice
	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewUserRealm(alice))
	users.Register(cross, "alice123")

	// Prop to delete user
	testing.SetOriginCaller(c)
	testing.SetRealm(testing.NewUserRealm(c))
	pr := users.ProposeDeleteUser(alice, "delete user test")
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

	data, _ := susers.ResolveName("alice123")
	if data == nil {
		println("Successfully deleted alice")
	}
}

// Output:
// --
// # GovDAO
// [> View Activity <](/r/gov/dao:activity)
//
// ## Members
// [> Go to Memberstore <](/r/gov/dao/v3/memberstore)
// ## Proposals
// ### [Prop #0 - User Registry V1: Delete user `alice123`](/r/gov/dao:0)
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// Status: ACTIVE
//
// Tiers eligible to vote: T1, T2, T3
//
// ---
//
//
// --
// ## Prop #0 - User Registry V1: Delete user `alice123`
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// delete user test
//
//
//
// ---
//
// ### Stats
// - **Proposal is open for votes**
// - Tiers eligible to vote: T1, T2, T3
// - YES PERCENT: 0%
// - NO PERCENT: 0%
//
// [Detailed voting list](/r/gov/dao:0/votes)
//
// ---
//
// ### Actions
// [Vote YES](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=YES&pid=0) | [Vote NO](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=NO&pid=0) | [Vote ABSTAIN](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=ABSTAIN&pid=0)
//
// WARNING: Please double check transaction data before voting.
// --
// --
// ## Prop #0 - User Registry V1: Delete user `alice123`
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// delete user test
//
//
//
// ---
//
// ### Stats
// - **Proposal is open for votes**
// - Tiers eligible to vote: T1, T2, T3
// - YES PERCENT: 100%
// - NO PERCENT: 0%
//
// [Detailed voting list](/r/gov/dao:0/votes)
//
// ---
//
// ### Actions
// [Vote YES](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=YES&pid=0) | [Vote NO](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=NO&pid=0) | [Vote ABSTAIN](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=ABSTAIN&pid=0)
//
// WARNING: Please double check transaction data before voting.
// --
// --
// ## Prop #0 - User Registry V1: Delete user `alice123`
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// delete user test
//
//
//
// ---
//
// ### Stats
// - **PROPOSAL HAS BEEN ACCEPTED**
// - Tiers eligible to vote: T1, T2, T3
// - YES PERCENT: 100%
// - NO PERCENT: 0%
//
// [Detailed voting list](/r/gov/dao:0/votes)
//
// ---
//
// ### Actions
// [Vote YES](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=YES&pid=0) | [Vote NO](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=NO&pid=0) | [Vote ABSTAIN](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=ABSTAIN&pid=0)
//
// WARNING: Please double check transaction data before voting.
// Successfully deleted alice
