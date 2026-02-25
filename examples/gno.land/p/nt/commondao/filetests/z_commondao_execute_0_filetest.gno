// PKGPATH: gno.land/r/test
package test

import (
	"chain/runtime"
	"testing"
	"time"

	"gno.land/p/nt/commondao"
)

const member address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

var (
	dao          *commondao.CommonDAO
	proposal     *commondao.Proposal
	executorPath string
)

type testPropDef struct{}

func (testPropDef) Title() string                               { return "" }
func (testPropDef) Body() string                                { return "" }
func (testPropDef) VotingPeriod() time.Duration                 { return 0 }
func (testPropDef) Tally(commondao.VotingContext) (bool, error) { return true, nil }

func (testPropDef) Executor() commondao.ExecFunc {
	return func(realm) error {
		executorPath = runtime.PreviousRealm().PkgPath()
		return nil
	}
}

func init() {
	dao = commondao.New(commondao.WithMember(member))
	proposal = dao.MustPropose(member, testPropDef{})
}

func main() {
	// Make sure proposal is executed in a realm different than the one where definition has been defined
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/testing/dao"))

	err := dao.Execute(proposal.ID())

	println(err == nil)
	println(string(proposal.Status()))
	println(executorPath)
}

// Output:
// true
// executed
// gno.land/r/testing/dao
