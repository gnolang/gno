// PKGPATH: gno.land/r/definition/test
package test

import (
	"chain/runtime"
	"errors"
	"time"

	"gno.land/p/nt/commondao"
	"gno.land/p/nt/commondao/exts/definition"
)

var dao = commondao.New()

// CreateMemberProposal creates a new example proposal to add a DAO member.
func CreateMemberProposal(member address) uint64 {
	if !member.IsValid() {
		panic("invalid member address")
	}

	// Define a function to validate that member doesn't exist within the DAO
	validate := func() error {
		if dao.Members().Has(member) {
			return errors.New("member already exists within the DAO")
		}
		return nil
	}

	// Define a custom tally function that approves proposals without votes
	tally := func(commondao.VotingContext) (bool, error) {
		return true, nil
	}

	// Define an executor to add the new member to the DAO
	executor := func(realm) error {
		dao.Members().Add(member)
		return nil
	}

	// Create a custom proposal definition for an example proposal type
	def := definition.MustNew(
		"Example Proposal",
		"This is a simple proposal example",
		definition.WithVotingPeriod(time.Hour*24*2), // 2 days
		definition.WithTally(tally),
		definition.WithValidation(validate),
		definition.WithExecutor(executor),
	)

	// Create a new proposal
	p := dao.MustPropose(runtime.PreviousRealm().Address(), def)
	return p.ID()
}

func main() {
	proposalID := CreateMemberProposal("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	println(proposalID)
}

// Output:
// 1
