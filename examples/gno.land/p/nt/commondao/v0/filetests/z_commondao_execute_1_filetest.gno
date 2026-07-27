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
func (testPropDef) VotingPeriod() time.Duration { return time.Hour }
func (testPropDef) Threshold() commondao.Threshold {
	return commondao.ThresholdSupermajority
}

func (testPropDef) Executor() commondao.ExecFunc {
	return func(realm) error {
		return errors.New("test error")
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
