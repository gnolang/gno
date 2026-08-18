// PKGPATH: gno.land/r/test
package test

import (
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

func (testPropDef) Title() string                  { return "" }
func (testPropDef) Body() string                   { return "" }
func (testPropDef) VotingPeriod() time.Duration    { return time.Hour } // Voting ends in 1 hour
func (testPropDef) Threshold() commondao.Threshold { return commondao.ThresholdSupermajority }

func (testPropDef) Executor() commondao.ExecFunc {
	return func(_ int, sub realm) error {
		executed = true
		return nil
	}
}

type testPropKind struct{}

func (testPropKind) Name() string { return "test" }
func (testPropKind) New(_ commondao.ReadonlyCommonDAO, _ any) (commondao.ProposalDefinition, error) {
	return testPropDef{}, nil
}

func init() {
	dao = commondao.New(
		commondao.WithCouncilMember(member),
		commondao.WithProposalKind(testPropKind{}),
	)
	var err error
	proposal, err = dao.Propose(member, "test", nil)
	if err != nil {
		panic(err)
	}
}

func main(cur realm) {
	// A supermajority of YES decides the proposal immediately
	err := dao.Vote(member, proposal.ID(), commondao.ChoiceYes, "")
	if err != nil {
		panic(err)
	}

	println(string(proposal.Status()))

	// Proposals decided early can be executed before the voting deadline.
	// The host mints the DAO-scoped sub and passes it into Execute.
	err = dao.Execute(proposal.ID(), cur.Sub("dao/1"))
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
