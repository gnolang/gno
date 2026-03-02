// PKGPATH: gno.land/r/test
package test

import (
	"testing"
	"time"

	"gno.land/p/nt/commondao/v0"
)

const member address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

var (
	dao      *commondao.CommonDAO
	proposal *commondao.Proposal
	executed bool
)

type testPropDef struct{}

func (testPropDef) Title() string                               { return "" }
func (testPropDef) Body() string                                { return "" }
func (testPropDef) VotingPeriod() time.Duration                 { return time.Hour } // Voting ends in 1 hour
func (testPropDef) Tally(commondao.VotingContext) (bool, error) { return true, nil }

func (testPropDef) Executor() commondao.ExecFunc {
	return func(realm) error {
		executed = true
		return nil
	}
}

func init() {
	dao = commondao.New(
		commondao.WithMember(member),
		commondao.DisableVotingDeadlineCheck(), // Disable to be able to execute before voting deadline
	)
	proposal = dao.MustPropose(member, testPropDef{})
}

func main() {
	// Make sure proposal is executed in a realm different than the one where definition has been defined
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/testing/dao"))

	// Should be able to execute before voting deadline because deadline check is disabled
	err := dao.Execute(proposal.ID())
	if err != nil {
		panic(err)
	}

	println(executed)
	println(string(proposal.Status()))
}

// Output:
// true
// executed
