package main

import (
	"std"
	"time"

	pcommondao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

type propDef struct{}

func (propDef) Title() string                                   { return "" }
func (propDef) Body() string                                    { return "" }
func (propDef) Quorum() float64                                 { return 0 }
func (propDef) VotingPeriod() time.Duration                     { return 0 }
func (propDef) Validate() error                                 { return nil }
func (propDef) Tally(pcommondao.ReadOnlyVotingRecord, int) bool { return false }
func (propDef) Execute() error                                  { return nil }

var (
	daoID    uint64
	proposal *pcommondao.Proposal
	vote     = pcommondao.ChoiceYes
)

func init() {
	// Use a realm caller
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))

	// Invite a user to be able to start creating DAOs
	std.TestSetOriginCaller(owner)
	commondao.Invite(user)

	// Create a new DAO which gives ownership to `commondao_test`
	std.TestSetOriginCaller(user)
	dao := commondao.New("Foo")
	daoID = dao.ID()

	// Configure DAO
	dao.Members().Add(user)

	// Create a new proposal
	proposal, _ = dao.Propose(user, propDef{})

	// User must be the caller to Vote()
	std.TestSetRealm(std.NewUserRealm(user))
}

func main() {
	commondao.Vote(daoID, proposal.ID(), vote)

	record := proposal.VotingRecord()
	if len(record.Votes()) != 1 {
		panic("expected a single vote")
	}

	println(record.HasVoted(user))
	println(record.Votes()[0].Choice == vote)
	println(record.Votes()[0].Choice == vote)
}

// Output:
// true
// true
// true
