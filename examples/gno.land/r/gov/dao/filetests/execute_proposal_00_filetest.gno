// PKGPATH: gno.land/r/test/execute
package execute

import (
	"errors"
	"testing"

	"gno.land/r/gov/dao"
	daoinit "gno.land/r/gov/dao/v3/init"
)

const user address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

var proposalID dao.ProposalID

func init() {
	// Initialize GovDAO members
	daoinit.InitWithUsers(user)

	// Create a proposal request that fails on execution
	cb := func(realm) error { return errors.New("Boom!") }
	request := dao.NewProposalRequest(
		"Test",
		"This proposal always fails on execution",
		dao.NewSimpleExecutor(cb, ""),
	)

	// Create the proposal
	testing.SetRealm(testing.NewUserRealm(user))
	proposalID = dao.MustCreateProposal(cross, request)

	// Vote on proposal
	vote := string(dao.YesVote)
	id := int64(proposalID)
	dao.MustVoteOnProposalSimple(cross, id, vote)
}

func main() {
	// Execute proposal, status should be REJECTED
	testing.SetRealm(testing.NewUserRealm(user))
	dao.ExecuteProposal(cross, proposalID, true)

	// Check proposal status
	println(dao.Render(proposalID.String()))
}

// Output:
// ## Prop #0 - Test
// Author: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
//
// This proposal always fails on execution
//
//
//
// ---
//
// ### Stats
// - **PROPOSAL HAS BEEN DENIED**
// REASON: execution failed: Boom!
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
