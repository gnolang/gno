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
func (testPropDef) Threshold() commondao.Threshold { return commondao.ThresholdSupermajority }

func (testPropDef) Executor() commondao.ExecFunc {
	return func(realm) error {
		executed = true
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
}

func main(cur realm) {
	// Make sure proposal is executed in a realm different than the one where definition has been defined
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/testing/dao"))

	// A supermajority of YES decides the proposal immediately
	err := dao.Vote(member, proposal.ID(), commondao.ChoiceYes, "")
	if err != nil {
		panic(err)
	}

	println(string(proposal.Status()))

	// Proposals decided early can be executed before the voting deadline
	err = dao.Execute(proposal.ID(), cur)
	if err != nil {
		panic(err)
	}

	println(executed)
	println(string(proposal.Status()))
}

// Output:
// passed
// true
// executed
