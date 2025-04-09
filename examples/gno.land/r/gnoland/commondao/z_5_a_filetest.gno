package main

import (
	"std"
	"testing"
	"time"

	pdao "gno.land/p/gnoland/commondao"

	"gno.land/r/gnoland/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

type propDef struct{}

func (propDef) Title() string                                                 { return "" }
func (propDef) Body() string                                                  { return "" }
func (propDef) VotingPeriod() time.Duration                                   { return 0 }
func (propDef) Validate() error                                               { return nil }
func (propDef) Tally(pdao.ReadonlyVotingRecord, pdao.MemberSet) (bool, error) { return false, nil }
func (propDef) Execute() error                                                { return nil }

var (
	daoID    uint64
	proposal *pdao.Proposal
	vote     pdao.VoteChoice = pdao.ChoiceYes
)

func init() {
	// Use a realm caller
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))

	// Invite a user to be able to start creating DAOs
	testing.SetOriginCaller(owner)
	commondao.Invite(user)

	// Create a new DAO which gives ownership to `commondao_test`
	testing.SetOriginCaller(user)
	dao := commondao.New("Foo")
	daoID = dao.ID()

	// Configure DAO
	dao.Members().Add(user)

	// Create a new proposal
	proposal, _ = dao.Propose(user, propDef{})

	// User must be the caller to Vote()
	testing.SetRealm(std.NewUserRealm(user))
}

func main() {
	commondao.Vote(daoID, proposal.ID(), vote, "")

	record := proposal.VotingRecord()
	if record.Size() != 1 {
		panic("expected a single vote")
	}

	println(record.HasVoted(user))
	record.Iterate(0, record.Size(), func(v pdao.Vote) bool {
		println(v.Choice == vote)
	})
}

// Output:
// true
// true
