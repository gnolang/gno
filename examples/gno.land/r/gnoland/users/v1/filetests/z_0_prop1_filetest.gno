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

	// Prop to change name
	testing.SetOriginCaller(c)
	testing.SetRealm(testing.NewUserRealm(c))
	pr := users.ProposeNewName(alice, "alice_new123")
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

	data, _ := susers.ResolveName("alice_new123")
	println(data.Addr())
}

// Output:
// --
// # Governance DAO
//
// **1** proposals total | **1** active | **0** completed | [Members](/r/gov/dao/v3/memberstore)
//
// ## Proposals
//
// ### [User Registry V1: Rename user `alice123` to `alice_new123`](/r/gov/dao:0)
//
// **Proposal #0** | ACTIVE | by [@1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm](/u/1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
//
// **YES:** 3 (0.0%) | **NO:** 3 (0.0%) | _Eligible: T1, T2, T3_
//
// ---
//
//
//
// --
// # Proposal #0 - User Registry V1: Rename user `alice123` to `alice_new123`
//
// [View All Votes](/r/gov/dao:0/votes) | [Back to Proposals](/r/gov/dao)
//
// **ACTIVE** - Voting in progress
//
// ## Details
//
// | | |
// |---|---|
// | **Author** | [@1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm](/u/1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm) |
// | **Status** | ACTIVE |
// | **Eligible Tiers** | T1, T2, T3 |
// | **YES** | 0.00% |
// | **NO** | 0.00% |
//
// ## Description
//
//
//
// ## Voting Statistics
//
// **Voting in Progress**
//
// | Vote Option | Votes | Percentage |
// |-------------|-------|------------|
// | YES | 3 | 0.00% |
// | NO | 3 | 0.00% |
//
// **Eligible to vote:** T1, T2, T3
//
// ## Actions
//
// **Cast your vote:**
//
// [Vote YES](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=YES&pid=0) | [Vote NO](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=NO&pid=0) | [Vote ABSTAIN](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=ABSTAIN&pid=0)
//
// > **Note:** Please verify all transaction details before submitting your vote.
//
// > **Requirement:** You must have a registered namespace to cast a vote.
// --
// --
// # Proposal #0 - User Registry V1: Rename user `alice123` to `alice_new123`
//
// [View All Votes](/r/gov/dao:0/votes) | [Back to Proposals](/r/gov/dao)
//
// **ACTIVE** - Voting in progress
//
// ## Details
//
// | | |
// |---|---|
// | **Author** | [@1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm](/u/1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm) |
// | **Status** | ACTIVE |
// | **Eligible Tiers** | T1, T2, T3 |
// | **YES** | 100.00% |
// | **NO** | 0.00% |
//
// ## Description
//
//
//
// ## Voting Statistics
//
// **Voting in Progress**
//
// | Vote Option | Votes | Percentage |
// |-------------|-------|------------|
// | YES | 3 | 100.00% |
// | NO | 3 | 0.00% |
//
// **Eligible to vote:** T1, T2, T3
//
// ## Actions
//
// **Cast your vote:**
//
// [Vote YES](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=YES&pid=0) | [Vote NO](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=NO&pid=0) | [Vote ABSTAIN](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=ABSTAIN&pid=0)
//
// > **Note:** Please verify all transaction details before submitting your vote.
//
// > **Requirement:** You must have a registered namespace to cast a vote.
// --
// --
// # Proposal #0 - User Registry V1: Rename user `alice123` to `alice_new123`
//
// [View All Votes](/r/gov/dao:0/votes) | [Back to Proposals](/r/gov/dao)
//
// **ACCEPTED** - This proposal has been approved and executed
//
// ## Details
//
// | | |
// |---|---|
// | **Author** | [@1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm](/u/1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm) |
// | **Status** | ACCEPTED |
// | **Eligible Tiers** | T1, T2, T3 |
//
// ## Description
//
//
//
// ## Voting Statistics
//
// **PROPOSAL ACCEPTED**
//
// | Vote Option | Votes | Percentage |
// |-------------|-------|------------|
// | YES | 3 | 100.00% |
// | NO | 3 | 0.00% |
//
// **Eligible to vote:** T1, T2, T3
//
// ## Actions
//
// > This proposal has been accepted and executed. No further actions needed.
//
// g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh
