package main

import (
	"std"
	"testing"
	"time"

	pdao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
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
)

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(std.NewUserRealm(owner))
	commondao.Invite(cross, user)

	// Create a new DAO which gives ownership to `test`
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test"))
	dao := commondao.New("Foo")
	daoID = dao.ID()

	// Configure DAO
	dao.Members().Add(user)

	// Create a new proposal
	proposal, _ = dao.Propose(user, propDef{})
}

func main() {
	testing.SetRealm(std.NewUserRealm(user))

	commondao.Vote(cross, daoID, proposal.ID(), "invalid", "")
}

// Error:
// invalid vote choice
