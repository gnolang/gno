// PKGPATH: gno.land/r/test
package test

import (
	"time"

	"gno.land/p/nt/commondao/v0"
)

const member address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

var (
	dao        *commondao.CommonDAO
	proposal   *commondao.Proposal
	subValid   bool
	subAddress address
)

type testPropDef struct{}

func (testPropDef) Title() string               { return "" }
func (testPropDef) Body() string                { return "" }
func (testPropDef) VotingPeriod() time.Duration { return time.Hour }
func (testPropDef) Threshold() commondao.Threshold {
	return commondao.ThresholdSupermajority
}

func (testPropDef) Executor() commondao.ExecFunc {
	// Executors are non-crossing: they must not rely on
	// unsafe.PreviousRealm() for realm identity (there is no crossing frame
	// for the executor). The DAO-scoped identity is the sub passed by the
	// host, which the executor can read and move value from.
	return func(_ int, sub realm) error {
		subValid = sub.Address().IsValid()
		subAddress = sub.Address()
		return nil
	}
}

type testPropKind struct{}

func (testPropKind) Name() string { return "test" }
func (testPropKind) New(_ *commondao.CommonDAO, _ any) (commondao.ProposalDefinition, error) {
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

	// A single YES from the one-member council passes the proposal early
	if err := dao.Vote(member, proposal.ID(), commondao.ChoiceYes, ""); err != nil {
		panic(err)
	}
}

func main(cur realm) {
	// The host mints the DAO-scoped sub and passes it into Execute; the
	// executor runs with that sub as its value-movement authority.
	sub := cur.Sub("dao/1")
	err := dao.Execute(proposal.ID(), sub)

	println(err == nil)
	println(string(proposal.Status()))
	println(subValid)
	println(subAddress == sub.Address())
}

// Output:
// true
// executed
// true
// true
