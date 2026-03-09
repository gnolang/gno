// PKGPATH: gno.land/r/test
package test

import (
	"errors"
	"time"

	"gno.land/p/nt/commondao/v0"
)

const member address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

var (
	dao      *commondao.CommonDAO
	proposal *commondao.Proposal
)

type testPropDef struct{}

func (testPropDef) Title() string                               { return "" }
func (testPropDef) Body() string                                { return "" }
func (testPropDef) VotingPeriod() time.Duration                 { return 0 }
func (testPropDef) Tally(commondao.VotingContext) (bool, error) { return true, nil }

func (testPropDef) Executor() commondao.ExecFunc {
	return func(realm) error {
		return errors.New("test error")
	}
}

func init() {
	dao = commondao.New(commondao.WithMember(member))
	proposal = dao.MustPropose(member, testPropDef{})
}

func main() {
	err := dao.Execute(proposal.ID())
	if err != nil {
		panic(err)
	}

	println(string(proposal.Status()))
	println(proposal.StatusReason())
}

// Output:
// failed
// test error
