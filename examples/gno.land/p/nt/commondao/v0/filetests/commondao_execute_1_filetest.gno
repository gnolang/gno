// PKGPATH: gno.land/r/test/commondao_execute
package commondao_execute

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

type mockPropDef struct{}

func (mockPropDef) Title() string                               { return "" }
func (mockPropDef) Body() string                                { return "" }
func (mockPropDef) VotingPeriod() time.Duration                 { return 0 }
func (mockPropDef) Tally(commondao.VotingContext) (bool, error) { return true, nil }

func (mockPropDef) Executor() commondao.ExecFunc {
	return func(string) error {
		return errors.New("test error")
	}
}

func init() {
	dao = commondao.New(commondao.WithMember(member))
	proposal = dao.MustPropose(member, mockPropDef{})
}

func main(cur realm) {
	err := dao.Execute(proposal.ID(), cur)
	if err != nil {
		panic(err)
	}

	println(string(proposal.Status()))
	println(proposal.StatusReason())
}

// Output:
// failed
// test error
