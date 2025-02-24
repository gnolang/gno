package main

import (
	"std"
	"time"

	pcommondao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
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
)

func init() {
	// Use a realm caller
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))

	// Invite a user to be able to start creating DAOs
	std.TestSetOrigCaller(owner)
	commondao.Invite(user)

	// Create a new DAO which gives ownership to `commondao_test`
	std.TestSetOrigCaller(user)
	dao := commondao.New("Foo")
	daoID = dao.ID()

	// Configure DAO
	dao.AddMember(user)

	// Create a new proposal
	proposal, _ = dao.Propose(user, propDef{})
}

func main() {
	commondao.Vote(daoID, 404, "invalid")
}

// Error:
// invalid vote choice
