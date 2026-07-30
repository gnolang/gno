// PKGPATH: gno.land/r/test
package test

import (
	"time"

	"gno.land/p/nt/commondao/v0"
)

const member address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

var (
	dao                  *commondao.CommonDAO
	proposal             *commondao.Proposal
	activeSizeDuringExec int
	execCount            int
)

// probePropDef observes, from inside its own executor, whether its proposal
// is still in active storage. Execute removes the proposal from active
// storage BEFORE running Validate/the executor (remove-before-run); that is
// exactly what makes a re-entrant Execute of the same proposal impossible
// (it would find nothing in active storage). So during the executor the
// active-proposal count must already be 0. Deleting that removal would
// leave the count at 1 — a re-entrancy hole.
type probePropDef struct{}

func (probePropDef) Title() string               { return "" }
func (probePropDef) Body() string                { return "" }
func (probePropDef) VotingPeriod() time.Duration { return time.Hour }
func (probePropDef) Threshold() commondao.Threshold {
	return commondao.ThresholdSupermajority
}

func (probePropDef) Executor() commondao.ExecFunc {
	return func(_ int, sub realm) error {
		execCount++
		activeSizeDuringExec = dao.ActiveProposalsSize()
		return nil
	}
}

type probePropKind struct{}

func (probePropKind) Name() string { return "probe" }
func (probePropKind) New(_ commondao.ReadonlyCommonDAO, _ any) (commondao.ProposalDefinition, error) {
	return probePropDef{}, nil
}

func init() {
	dao = commondao.New(
		commondao.WithCouncilMember(member),
		commondao.WithProposalKind(probePropKind{}),
	)
	var err error
	proposal, err = dao.Propose(member, "probe", nil)
	if err != nil {
		panic(err)
	}
	// A single YES from the one-member council passes the proposal early.
	if err := dao.Vote(member, proposal.ID(), commondao.ChoiceYes, ""); err != nil {
		panic(err)
	}
}

func main(cur realm) {
	err := dao.Execute(proposal.ID(), cur.Sub("dao/1"))

	println("outer err nil:", err == nil)
	println("status:", string(proposal.Status()))
	println("executor ran once:", execCount == 1)
	println("active count during executor:", activeSizeDuringExec)
}

// Output:
// outer err nil: true
// status: executed
// executor ran once: true
// active count during executor: 0
