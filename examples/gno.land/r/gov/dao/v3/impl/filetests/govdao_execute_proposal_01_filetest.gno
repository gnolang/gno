// PKGPATH: gno.land/r/test/govdao
package govdao

import (
	"strconv"
	"testing"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/memberstore"
)

const user address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

var (
	executor   dao.Executor
	proposalID dao.ProposalID
	renderPath string
	govdao     = impl.NewGovDAO()
)

func init() {
	// Initialize GovDAO members
	memberstore.Get().DeleteAll()
	memberstore.Get().SetTier(memberstore.T1)
	memberstore.Get().SetMember(memberstore.T1, user, &memberstore.Member{InvitationPoints: 3})
	dao.UpdateImpl(cross, dao.UpdateRequest{DAO: impl.NewGovDAO()})

	// Create a dummy executor
	cb := func(realm) error { return nil }
	executor = dao.NewSimpleExecutor(cb, "")

	// Create a proposal request that pass on execution
	request := dao.NewProposalRequest("Test", "This proposal always pass on execution", executor)

	// Create the proposal from a realm so GovDAO instance is able to render the proposal
	testing.SetRealm(testing.NewUserRealm(user))
	proposalID = dao.MustCreateProposal(cross, request)
	renderPath = strconv.FormatUint(uint64(proposalID), 10)

	// Register proposal with the local GovDAO instance
	govdao.PostCreateProposal(request, proposalID)
}

func main() {
	// Execute proposal, status should be ACTIVE
	err := govdao.ExecuteProposal(proposalID, executor)

	println(err == nil)
	println()
	println(govdao.Render("gno.land/r/gov/dao/v3/impl", renderPath))
}

// Output:
// true
//
// ## Prop #0 - Test
// Author: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
//
// This proposal always pass on execution
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
// [Detailed voting list](/r/gov/dao/v3/impl:0/votes)
//
// ---
//
// ### Actions
// [Vote YES](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=YES&pid=0) | [Vote NO](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=NO&pid=0) | [Vote ABSTAIN](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=ABSTAIN&pid=0)
//
// WARNING: Please double check transaction data before voting.
