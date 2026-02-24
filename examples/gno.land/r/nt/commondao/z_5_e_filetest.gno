package main

import (
	"testing"
	"time"

	pdao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

type propDef struct{}

func (propDef) Title() string                          { return "" }
func (propDef) Body() string                           { return "" }
func (propDef) VotingPeriod() time.Duration            { return time.Hour }
func (propDef) Validate() error                        { return nil }
func (propDef) Tally(pdao.VotingContext) (bool, error) { return false, nil }

var (
	daoID    uint64
	proposal *pdao.Proposal
)

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross, user)

	// Create a new DAO which gives ownership to `test`
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))
	dao := commondao.New(cross, "Foo")
	daoID = dao.ID()

	// Configure DAO
	dao.Members().Add(user)

	// Create a new proposal
	proposal, _ = dao.Propose(user, propDef{})
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.Vote(cross, daoID, proposal.ID(), "invalid", "")
}

// Error:
// invalid vote choice
