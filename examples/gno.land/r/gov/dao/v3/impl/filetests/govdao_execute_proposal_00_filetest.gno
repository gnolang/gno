// PKGPATH: gno.land/r/test/govdao
package govdao

import (
	"errors"
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

	// Create an executor that always fails
	cb := func(realm) error { return errors.New("Boom!") }
	executor = dao.NewSimpleExecutor(cb, "")

	// Create a proposal request that fails on execution
	request := dao.NewProposalRequest("Test", "This proposal always fails on execution", executor)

	// Create the proposal from a realm so GovDAO instance is able to render the proposal
	testing.SetRealm(testing.NewUserRealm(user))
	proposalID = dao.MustCreateProposal(cross, request)
	renderPath = strconv.FormatUint(uint64(proposalID), 10)

	// Register proposal with the local GovDAO instance
	govdao.PostCreateProposal(request, proposalID)
}

func main() {
	// Execute proposal, status should be REJECTED
	err := govdao.ExecuteProposal(proposalID, executor)

	println(err.Error())
	println()
	println(govdao.Render("gno.land/r/gov/dao/v3/impl", renderPath))
}

// Output:
// Boom!
//
// # Proposal #0 - Test
//
// [View All Votes](/r/gov/dao/v3/impl:0/votes) | [Back to Proposals](/r/gov/dao/v3/impl)
//
// **REJECTED** - This proposal has been denied
//
// ## Details
//
// | | |
// |---|---|
// | **Author** | g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 |
// | **Status** | REJECTED _(Reason: execution failed: Boom!)_ |
// | **Eligible Tiers** | T1, T2, T3 |
//
// ## Description
//
// This proposal always fails on execution
//
// ## Voting Statistics
//
// **PROPOSAL REJECTED**
//
// > **Reason:** execution failed: Boom!
//
// | Vote Option | Votes | Percentage |
// |-------------|-------|------------|
// | YES | 3 | 0.00% |
// | NO | 3 | 0.00% |
// | ABSTAIN | 3 | 0.00% |
//
// **Eligible to vote:** T1, T2, T3
//
// ## Actions
//
// > This proposal has been rejected. No further actions available.
