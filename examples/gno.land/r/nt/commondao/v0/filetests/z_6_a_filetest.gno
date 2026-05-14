// PKGPATH: gno.land/r/demo/test
package test

import (
	"testing"
	"time"

	pdao "gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/testutils/v0"

	"gno.land/r/nt/commondao/v0"
)

const owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	dao      *pdao.CommonDAO
	proposal *pdao.Proposal
	executed bool
	user1    = testutils.TestAddress("user1")
)

type propDef struct{}

func (propDef) Title() string                          { return "" }
func (propDef) Body() string                           { return "" }
func (propDef) VotingPeriod() time.Duration            { return 0 }
func (propDef) Validate() error                        { return nil }
func (propDef) Tally(pdao.VotingContext) (bool, error) { return true, nil }

func (propDef) Executor() pdao.ExecFunc {
	return func(realm) error {
		executed = true
		return nil
	}
}

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross, user1)

	// Create a new DAO which gives ownership to `test`
	testing.SetRealm(testing.NewUserRealm(user1))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	dao = commondao.New(cross, "Foo")

	// Configure DAO
	dao.Members().Add(user1)

	// Create a new proposal
	proposal, _ = dao.Propose(user1, propDef{})
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user1))

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
