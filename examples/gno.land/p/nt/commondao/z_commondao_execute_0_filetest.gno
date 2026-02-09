// PKGPATH: gno.land/r/test
package test

import (
	"time"

	"gno.land/p/nt/commondao"
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
func (testPropDef) Execute(cur realm) error                     { return nil }
func (testPropDef) Tally(commondao.VotingContext) (bool, error) { return true, nil }

func init() {
	dao = commondao.New(commondao.WithMember(member))
	proposal = dao.MustPropose(member, testPropDef{})
}

func main() {
	err := dao.Execute(proposal.ID())

	println(err == nil)
	println(string(proposal.Status()))
}

// Output:
// true
// executed
