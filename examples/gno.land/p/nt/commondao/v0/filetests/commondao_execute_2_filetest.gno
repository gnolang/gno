// PKGPATH: gno.land/r/test/commondao_execute
package commondao_execute

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

type mockPropDef struct{}

func (mockPropDef) Title() string                               { return "" }
func (mockPropDef) Body() string                                { return "" }
func (mockPropDef) VotingPeriod() time.Duration                 { return time.Hour } // Voting ends in 1 hour
func (mockPropDef) Tally(commondao.VotingContext) (bool, error) { return true, nil }

func (mockPropDef) Executor() commondao.ExecFunc {
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
	proposal = dao.MustPropose(member, mockPropDef{})
}

func main(cur realm) {
	// Make sure proposal is executed in a realm different than the one where definition has been defined
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/testing/dao"))

	// Should be able to execute before voting deadline because deadline check is disabled
	err := dao.Execute(proposal.ID(), cur)
	if err != nil {
		panic(err)
	}

	println(executed)
	println(string(proposal.Status()))
}

// Output:
// true
// executed
