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
	// Execute proposal, execution should fail
	testing.SetRealm(testing.NewUserRealm(user))
	dao.ExecuteProposal(cross, proposalID, false)
}

// Error:
// Boom!
