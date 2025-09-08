// PKGPATH: gno.land/r/demo/test
package test

import (
	"std"
	"testing"
	"time"

	pdao "gno.land/p/nt/commondao"
	"gno.land/p/nt/testutils"

	"gno.land/r/nt/commondao"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	dao      *pdao.CommonDAO
	proposal *pdao.Proposal
	executed bool
	user1    = testutils.TestAddress("user1")
)

type propDef struct{}

func (propDef) Title() string                                                 { return "" }
func (propDef) Body() string                                                  { return "" }
func (propDef) VotingPeriod() time.Duration                                   { return 0 }
func (propDef) Validate() error                                               { return nil }
func (propDef) Tally(pdao.ReadonlyVotingRecord, pdao.MemberSet) (bool, error) { return true, nil }

func (propDef) Execute(cur realm) error {
	executed = true
	return nil
}

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(std.NewUserRealm(owner))
	commondao.Invite(cross, user1)

	// Create a new DAO which gives ownership to `test`
	testing.SetRealm(std.NewUserRealm(user1))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/test"))
	dao = commondao.New(cross, "Foo")

	// Configure DAO
	dao.Members().Add(user1)

	// Create a new proposal
	proposal, _ = dao.Propose(user1, propDef{})
}

func main() {
	testing.SetRealm(std.NewUserRealm(user1))

	commondao.Execute(cross, dao.ID(), proposal.ID())

	p := dao.FinishedProposals().Get(proposal.ID())
	if p == nil {
		panic("expected proposal to be finished")
	}

	println(string(p.Status()))
	println(executed)
}

// Output:
// executed
// true
