// PKGPATH: gno.land/r/test/commondao_execute
package commondao_execute

import (
	"testing"
	"time"

	"gno.land/p/nt/commondao/v0"
)

const member address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

var (
	dao          *commondao.CommonDAO
	proposal     *commondao.Proposal
	executorPath string
)

type mockPropDef struct{}

func (mockPropDef) Title() string                               { return "" }
func (mockPropDef) Body() string                                { return "" }
func (mockPropDef) VotingPeriod() time.Duration                 { return 0 }
func (mockPropDef) Tally(commondao.VotingContext) (bool, error) { return true, nil }

func (mockPropDef) Executor() commondao.ExecFunc {
	return func(pkgPath string) error {
		executorPath = pkgPath
		return nil
	}
}

func init() {
	dao = commondao.New(commondao.WithMember(member))
	proposal = dao.MustPropose(member, mockPropDef{})
}

func main(cur realm) {
	// Make sure proposal is executed in a realm different than the one where definition has been defined
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/testing/dao"))

	err := dao.Execute(proposal.ID(), cur)

	println(err == nil)
	println(string(proposal.Status()))
	println(executorPath)
}

// Output:
// true
// executed
// gno.land/r/testing/dao
