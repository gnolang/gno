// PKGPATH: gno.land/r/test
package test

import (
	"chain/runtime/unsafe"
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

type testPropDef struct{}

func (testPropDef) Title() string               { return "" }
func (testPropDef) Body() string                { return "" }
func (testPropDef) VotingPeriod() time.Duration { return time.Hour }
func (testPropDef) Threshold() commondao.Threshold {
	return commondao.ThresholdSupermajority
}

func (testPropDef) Executor() commondao.ExecFunc {
	return func(realm) error {
		executorPath = unsafe.PreviousRealm().PkgPath()
		return nil
	}
}

func init() {
	dao = commondao.New(commondao.WithCouncilMember(member))
	var err error
	proposal, err = dao.Propose(member, testPropDef{})
	if err != nil {
		panic(err)
	}

	// A single YES from the one-member council passes the proposal early
	if err := dao.Vote(member, proposal.ID(), commondao.ChoiceYes, ""); err != nil {
		panic(err)
	}
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
