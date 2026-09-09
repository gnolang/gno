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

func init(cur realm) {
	// Initialize GovDAO members
	memberstore.Get(0, cur).DeleteAll()
	memberstore.Get(0, cur).SetTier(memberstore.T1)
	memberstore.Get(0, cur).SetMember(memberstore.T1, user, memberstore.NewMember(3))
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), nil))

	// Create an executor that always fails
	cb := func(realm) error { return errors.New("Boom!") }
	executor = dao.NewSimpleExecutor(0, cur, cb, "")

	// Create a proposal request that fails on execution
	request := dao.NewProposalRequest("Test", "This proposal always fails on execution", executor)

	// Create the proposal from a realm so GovDAO instance is able to render the proposal
	testing.SetRealm(testing.NewUserRealm(user))
	proposalID = dao.MustCreateProposal(cross(cur), request)
	renderPath = strconv.FormatUint(uint64(proposalID), 10)

	// Register proposal with the local GovDAO instance
	govdao.PostCreateProposal(0, cur, request, proposalID)
}

func main(cur realm) {
	// Execute proposal, status should be REJECTED
	err := govdao.ExecuteProposal(0, cur, proposalID, executor)

	println(err.Error())
	println()
	println(govdao.Render(cross(cur), "gno.land/r/gov/dao/v3/impl", renderPath))
}

// Output:
// Boom!
//
// ## Prop #0 - Test
// Author: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5\
// Executes: [`gno.land/r/test/govdao`](/r/test/govdao$source)\
// Status: REJECTED, execution failed: Boom\!\
// Eligible: T1, T2, T3\
// Votes: YES 0.00%, NO 0.00%, ABSTAIN 0.00% ([detail](/r/gov/dao/v3/impl:0/votes))
//
// ---
//
// ### Description
//
// This proposal always fails on execution
//
// ---
//
// ### Actions
// [Vote YES](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=YES&pid=0) | [Vote NO](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=NO&pid=0) | [Vote ABSTAIN](/r/gov/dao$help&func=MustVoteOnProposalSimple&option=ABSTAIN&pid=0)
//
// WARNING: Please double check transaction data before voting.
