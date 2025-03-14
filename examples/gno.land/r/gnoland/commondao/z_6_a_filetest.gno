package main

import (
	"std"
	"time"

	"gno.land/p/demo/testutils"
	pdao "gno.land/p/gnoland/commondao"

	"gno.land/r/gnoland/commondao"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

type propDef struct{}

func (propDef) Title() string                                                 { return "" }
func (propDef) Body() string                                                  { return "" }
func (propDef) VotingPeriod() time.Duration                                   { return 0 }
func (propDef) Validate() error                                               { return nil }
func (propDef) Tally(pdao.ReadonlyVotingRecord, pdao.MemberSet) (bool, error) { return true, nil }

func (propDef) Execute() error {
	executed = true
	return nil
}

var (
	dao      *pdao.CommonDAO
	proposal *pdao.Proposal
	executed bool
	user1    = testutils.TestAddress("user1")
	user2    = testutils.TestAddress("user2")
	user3    = testutils.TestAddress("user3")
)

func init() {
	// Use a realm caller
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))

	// Invite a user to be able to start creating DAOs
	std.TestSetOriginCaller(owner)
	commondao.Invite(user1)

	// Create a new DAO which gives ownership to `commondao_test`
	std.TestSetOriginCaller(user1)
	dao = commondao.New("Foo")

	// Configure DAO
	dao.Members().Add(user1)
	dao.Members().Add(user2)
	dao.Members().Add(user3)

	// Create a new proposal
	proposal, _ = dao.Propose(user1, propDef{})

	// Submit user1 vote
	commondao.Vote(dao.ID(), proposal.ID(), pdao.ChoiceYes, "")

	// Submit user2 vote
	std.TestSetOriginCaller(user2)
	commondao.Vote(dao.ID(), proposal.ID(), pdao.ChoiceYes, "")
}

func main() {
	commondao.Execute(dao.ID(), proposal.ID())

	p, found := dao.GetFinishedProposal(proposal.ID())
	if !found {
		panic("expected proposal to be finished")
	}

	println(p.Status() == pdao.StatusPassed)
	println(executed)
}

// Output:
// true
// true
