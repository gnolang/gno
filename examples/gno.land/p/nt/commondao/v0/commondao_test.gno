package commondao_test

import (
	"errors"
	"testing"
	"time"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/p/nt/commondao/v0"
)

const (
	memberA address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	memberB address = "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"
	memberC address = "g1w4ek2u3jta047h6lta047h6lta047h6l9huexc"
	memberD address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
	memberE address = "g1yx35a8f4lhhhrarljrgr37u55gkqdrl0949ycf"
	memberF address = "g1cyh9qm4y43w38gfk77qxczrhqj2asdr8l7uk54"
	memberG address = "g1x95kxqpwvqjf98qpkvsay7tppkh6zs9fr7wwda"
	memberH address = "g12m7ryva2vsw3twqdtuc9awufwzwtk5al4v4ksd"
	memberI address = "g1manfred47kzduec920z88wfr64ylksmdcedlf5"
)

// probeKind is a pass-through proposal kind for tests: args is the
// proposal definition itself.
type probeKind struct{}

func (probeKind) Name() string { return "probe" }
func (probeKind) New(_ commondao.ReadonlyCommonDAO, args any) (commondao.ProposalDefinition, error) {
	d, _ := args.(commondao.ProposalDefinition)
	return d, nil
}

// namedKind is a probe kind with a configurable name.
type namedKind string

func (k namedKind) Name() string { return string(k) }
func (namedKind) New(_ commondao.ReadonlyCommonDAO, _ any) (commondao.ProposalDefinition, error) {
	return propDef{period: time.Hour}, nil
}

// failingKind is a probe kind whose factory always fails.
type failingKind struct{ err error }

func (failingKind) Name() string { return "failing" }
func (k failingKind) New(_ commondao.ReadonlyCommonDAO, _ any) (commondao.ProposalDefinition, error) {
	return nil, k.err
}

// hostKind is a probe kind that records the ID of the readonly host view
// passed to its factory. New receives only a ReadonlyCommonDAO (no mutable
// handle), so the kind can identify the host but cannot mutate it.
type hostKind struct{ hostID *uint64 }

func (hostKind) Name() string { return "host" }
func (k hostKind) New(dao commondao.ReadonlyCommonDAO, _ any) (commondao.ProposalDefinition, error) {
	*k.hostID = dao.ID()
	return propDef{period: time.Hour}, nil
}

// propDef is a base proposal definition decided by the default rules.
type propDef struct {
	title  string
	body   string
	period time.Duration
}

func (d propDef) Title() string               { return d.title }
func (d propDef) Body() string                { return d.body }
func (d propDef) VotingPeriod() time.Duration { return d.period }
func (propDef) Threshold() commondao.Threshold {
	return commondao.ThresholdSupermajority
}

// mustPropose creates a proposal through the probe kind or fails the test.
func mustPropose(t *testing.T, dao *commondao.CommonDAO, creator address, d commondao.ProposalDefinition) *commondao.Proposal {
	t.Helper()
	p, err := dao.Propose(creator, "probe", d)
	urequire.NoError(t, err, "propose")
	return p
}

// defaultPropDef is decided by the constitution's default rules.
type defaultPropDef struct {
	period    time.Duration
	threshold commondao.Threshold
	executed  *bool
	execErr   error
}

func (defaultPropDef) Title() string                 { return "Default" }
func (defaultPropDef) Body() string                  { return "" }
func (d defaultPropDef) VotingPeriod() time.Duration { return d.period }
func (d defaultPropDef) Threshold() commondao.Threshold {
	return d.threshold
}

func (d defaultPropDef) Executor() commondao.ExecFunc {
	return func(_ int, sub realm) error {
		if d.executed != nil {
			*d.executed = true
		}
		return d.execErr
	}
}

// exemptPropDef is exempt from the active proposals cap.
type exemptPropDef struct{ defaultPropDef }

func (exemptPropDef) CapExempt() {}

// execFnPropDef executes an arbitrary function on approval.
type execFnPropDef struct {
	propDef
	fn commondao.ExecFunc
}

func (d execFnPropDef) Executor() commondao.ExecFunc { return d.fn }

// validatingExecPropDef is a default rule definition with state validation.
type validatingExecPropDef struct {
	defaultPropDef
	err *error
}

func (d validatingExecPropDef) Validate() error { return *d.err }

func TestNew(t *testing.T) {
	dao := commondao.New()
	uassert.Equal(t, uint64(0), dao.ID())
	uassert.Equal(t, "", dao.Name())
	uassert.Equal(t, 0, dao.Council().Size())
	uassert.Equal(t, 0, dao.ActiveProposalsSize())
	uassert.Equal(t, 0, dao.FinishedProposalsSize())
	uassert.Equal(t, 0, len(dao.KindNames()))
	uassert.False(t, dao.IsDeleted())
	urequire.True(t, dao.Parent() == nil, "expect no parent")

	parent := commondao.New(commondao.WithName("root"))
	dao = commondao.New(
		commondao.WithID(7),
		commondao.WithName("child"),
		commondao.WithDescription("test"),
		commondao.WithParent(parent),
		commondao.WithCouncilMember(memberA),
		commondao.WithCouncilMember(memberB),
	)
	uassert.Equal(t, uint64(7), dao.ID())
	uassert.Equal(t, "child", dao.Name())
	uassert.Equal(t, "test", dao.Description())
	uassert.Equal(t, 2, dao.Council().Size())
	uassert.True(t, dao.Council().Has(memberA))
	uassert.Equal(t, "root", dao.Parent().Name())
}

func TestUpdateCouncil(t *testing.T) {
	dao := commondao.New(commondao.WithCouncilMember(memberA), commondao.WithProposalKind(probeKind{}))

	// Idempotent set semantics
	urequire.NoError(t, dao.UpdateCouncil([]address{memberB, memberB}, nil))
	uassert.Equal(t, 2, dao.Council().Size())
	urequire.NoError(t, dao.UpdateCouncil(nil, []address{memberC})) // absent remove
	uassert.Equal(t, 2, dao.Council().Size())

	// Overlapping add/remove is rejected
	err := dao.UpdateCouncil([]address{memberC}, []address{memberC})
	uassert.ErrorIs(t, err, commondao.ErrCouncilUpdateOverlap)

	// Emptying a non-empty council is rejected
	err = dao.UpdateCouncil(nil, []address{memberA, memberB})
	uassert.ErrorIs(t, err, commondao.ErrEmptyCouncil)
	uassert.Equal(t, 2, dao.Council().Size())

	// Removing all but one is allowed
	urequire.NoError(t, dao.UpdateCouncil(nil, []address{memberA}))
	uassert.Equal(t, 1, dao.Council().Size())
	uassert.True(t, dao.Council().Has(memberB))

	// Full council replacement in a single update is legal
	urequire.NoError(t, dao.UpdateCouncil([]address{memberC, memberD}, []address{memberB}))
	uassert.Equal(t, 2, dao.Council().Size())
	uassert.False(t, dao.Council().Has(memberB))
	uassert.True(t, dao.Council().Has(memberC))
	uassert.True(t, dao.Council().Has(memberD))

	// More removals than members is fine as long as a member survives
	// (the emptiness check is by identity, not by count)
	urequire.NoError(t, dao.UpdateCouncil(nil, []address{memberE, memberF, memberG}))
	uassert.Equal(t, 2, dao.Council().Size())

	// An empty council can be seeded
	empty := commondao.New()
	urequire.NoError(t, empty.UpdateCouncil(nil, nil))
	urequire.NoError(t, empty.UpdateCouncil([]address{memberA}, nil))
	uassert.Equal(t, 1, empty.Council().Size())
}

func TestPropose(t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithProposalKind(probeKind{}),
	)

	p, err := dao.Propose(memberA, "probe", propDef{period: time.Hour})
	urequire.NoError(t, err)
	uassert.Equal(t, uint64(1), p.ID())
	uassert.Equal(t, 1, dao.ActiveProposalsSize())
	urequire.True(t, dao.GetProposal(p.ID()) != nil, "expect proposal to be found")

	// The electorate is the council snapshot taken at proposal creation
	uassert.Equal(t, 1, p.Electorate().Size())
	urequire.NoError(t, dao.UpdateCouncil([]address{memberB}, nil))
	uassert.Equal(t, 1, p.Electorate().Size())
	uassert.False(t, p.Electorate().Has(memberB))

	// Definitions are required
	_, err = dao.Propose(memberA, "probe", nil)
	uassert.ErrorIs(t, err, commondao.ErrProposalDefinitionRequired)

	// Deleted DAOs reject proposals
	deleted := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithProposalKind(probeKind{}),
	)
	deleted.Dissolve("test")
	_, err = deleted.Propose(memberA, "probe", propDef{})
	uassert.ErrorIs(t, err, commondao.ErrDAOIsDeleted)
}

func TestProposeCap(t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithProposalKind(probeKind{}),
	)
	dao.SetMaxActiveProposals(2)

	_, err := dao.Propose(memberA, "probe", propDef{period: time.Hour})
	urequire.NoError(t, err)
	_, err = dao.Propose(memberA, "probe", propDef{period: time.Hour})
	urequire.NoError(t, err)

	// The cap rejects further proposals
	_, err = dao.Propose(memberA, "probe", propDef{period: time.Hour})
	uassert.ErrorIs(t, err, commondao.ErrMaxActiveProposals)

	// Cap exempt definitions bypass the cap
	_, err = dao.Propose(memberA, "probe", exemptPropDef{defaultPropDef{period: time.Hour}})
	urequire.NoError(t, err)

	// But are bounded to one active exempt proposal per creator
	_, err = dao.Propose(memberA, "probe", exemptPropDef{defaultPropDef{period: time.Hour}})
	uassert.ErrorIs(t, err, commondao.ErrMaxCapExemptProposals)
	_, err = dao.Propose(memberB, "probe", exemptPropDef{defaultPropDef{period: time.Hour}})
	urequire.NoError(t, err)
}

func TestProposeCapSlotFrees(t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithCouncilMember(memberB),
		commondao.WithCouncilMember(memberC),
		commondao.WithProposalKind(probeKind{}),
	)
	dao.SetMaxActiveProposals(1)

	p := mustPropose(t, dao, memberA, defaultPropDef{period: time.Hour})
	_, err := dao.Propose(memberA, "probe", defaultPropDef{period: time.Hour})
	urequire.ErrorIs(t, err, commondao.ErrMaxActiveProposals)

	// An immediate NO majority dismissal frees the slot
	urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceNo, ""))
	urequire.NoError(t, dao.Vote(memberB, p.ID(), commondao.ChoiceNo, ""))
	urequire.Equal(t, string(commondao.StatusDismissed), string(p.Status()))

	_, err = dao.Propose(memberA, "probe", defaultPropDef{period: time.Hour})
	urequire.NoError(t, err)
}

func TestDissolve(t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithProposalKind(probeKind{}),
	)

	// An early passed (but unexecuted) proposal is still in-flight
	p := mustPropose(t, dao, memberA, defaultPropDef{period: time.Hour})
	urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
	urequire.Equal(t, string(commondao.StatusPassed), string(p.Status()))

	dao.Dissolve("gone")

	uassert.True(t, dao.IsDeleted())
	uassert.Equal(t, 0, dao.ActiveProposalsSize())
	uassert.Equal(t, 1, dao.FinishedProposalsSize())
	uassert.Equal(t, string(commondao.StatusDismissed), string(p.Status()))
	uassert.Equal(t, "gone", p.StatusReason())
}

func TestVote(t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithCouncilMember(memberB),
		commondao.WithCouncilMember(memberC),
		commondao.WithProposalKind(probeKind{}),
	)
	p := mustPropose(t, dao, memberA, propDef{period: time.Hour})

	// Only electorate members can vote
	err := dao.Vote(memberD, p.ID(), commondao.ChoiceYes, "")
	uassert.ErrorIs(t, err, commondao.ErrNotElectorateMember)

	// Members added after proposal creation are not in the electorate
	urequire.NoError(t, dao.UpdateCouncil([]address{memberD}, nil))
	err = dao.Vote(memberD, p.ID(), commondao.ChoiceYes, "")
	uassert.ErrorIs(t, err, commondao.ErrNotElectorateMember)

	// Members removed after proposal creation remain in the electorate
	urequire.NoError(t, dao.UpdateCouncil(nil, []address{memberC}))
	urequire.NoError(t, dao.Vote(memberC, p.ID(), commondao.ChoiceAbstain, ""))

	// Invalid choices are rejected
	err = dao.Vote(memberA, p.ID(), commondao.VoteChoice("FOO"), "")
	uassert.ErrorIs(t, err, commondao.ErrInvalidVoteChoice)

	// Members can vote and change their vote
	urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
	urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceNo, ""))
	uassert.Equal(t, 0, p.VotingRecord().VoteCount(commondao.ChoiceYes))
	uassert.Equal(t, 1, p.VotingRecord().VoteCount(commondao.ChoiceNo))

	// Unknown proposals and deadlines are rejected
	err = dao.Vote(memberA, 404, commondao.ChoiceYes, "")
	uassert.ErrorIs(t, err, commondao.ErrProposalNotFound)

	expired := mustPropose(t, dao, memberA, propDef{}) // zero voting period
	err = dao.Vote(memberA, expired.ID(), commondao.ChoiceYes, "")
	uassert.ErrorIs(t, err, commondao.ErrVotingDeadlinePassed)

	// Deleted DAOs reject votes
	dao.Dissolve("test")
	err = dao.Vote(memberA, p.ID(), commondao.ChoiceYes, "")
	uassert.ErrorIs(t, err, commondao.ErrDAOIsDeleted)
}

func TestVoteEarlyTermination(t *testing.T) {
	newDAO := func() *commondao.CommonDAO {
		return commondao.New(
			commondao.WithCouncilMember(memberA),
			commondao.WithCouncilMember(memberB),
			commondao.WithCouncilMember(memberC),
			commondao.WithProposalKind(probeKind{}),
		)
	}

	t.Run("supermajority YES passes immediately", func(t *testing.T) {
		dao := newDAO()
		p := mustPropose(t, dao, memberA, defaultPropDef{period: time.Hour})

		urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
		uassert.Equal(t, string(commondao.StatusActive), string(p.Status()))

		urequire.NoError(t, dao.Vote(memberB, p.ID(), commondao.ChoiceYes, ""))
		uassert.Equal(t, string(commondao.StatusPassed), string(p.Status()))

		// Passed proposals stay active (pending execution) but reject votes
		uassert.Equal(t, 1, dao.ActiveProposalsSize())
		err := dao.Vote(memberC, p.ID(), commondao.ChoiceNo, "")
		uassert.ErrorIs(t, err, commondao.ErrStatusIsNotActive)
	})

	t.Run("NO majority dismisses immediately", func(t *testing.T) {
		dao := newDAO()
		p := mustPropose(t, dao, memberA, defaultPropDef{period: time.Hour})

		urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceNo, ""))
		urequire.NoError(t, dao.Vote(memberB, p.ID(), commondao.ChoiceNo, ""))

		uassert.Equal(t, string(commondao.StatusDismissed), string(p.Status()))
		uassert.Equal(t, 0, dao.ActiveProposalsSize())
		uassert.Equal(t, 1, dao.FinishedProposalsSize())
	})

	t.Run("abstain flip can trigger a pass", func(t *testing.T) {
		dao := commondao.New(
			commondao.WithCouncilMember(memberA),
			commondao.WithCouncilMember(memberB),
			commondao.WithCouncilMember(memberC),
			commondao.WithCouncilMember(memberD),
			commondao.WithCouncilMember(memberE),
			commondao.WithProposalKind(probeKind{}),
		)
		p := mustPropose(t, dao, memberA, defaultPropDef{period: time.Hour})

		urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
		urequire.NoError(t, dao.Vote(memberB, p.ID(), commondao.ChoiceYes, ""))
		urequire.NoError(t, dao.Vote(memberC, p.ID(), commondao.ChoiceYes, ""))
		urequire.NoError(t, dao.Vote(memberD, p.ID(), commondao.ChoiceNo, ""))
		urequire.NoError(t, dao.Vote(memberE, p.ID(), commondao.ChoiceNo, ""))
		uassert.Equal(t, string(commondao.StatusActive), string(p.Status())) // 9 < 10

		// NO -> ABSTAIN shrinks the denominator: D=4, 3*3 >= 2*4
		urequire.NoError(t, dao.Vote(memberE, p.ID(), commondao.ChoiceAbstain, ""))
		uassert.Equal(t, string(commondao.StatusPassed), string(p.Status()))
	})
}

func TestExpectedOutcome(t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithCouncilMember(memberB),
		commondao.WithCouncilMember(memberC),
		commondao.WithProposalKind(probeKind{}),
	)

	// Default rule proposals report the live tally
	p := mustPropose(t, dao, memberA, defaultPropDef{period: time.Hour})
	uassert.Equal(t, int(commondao.OutcomePending), int(p.ExpectedOutcome()))
	urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
	uassert.Equal(t, int(commondao.OutcomePending), int(p.ExpectedOutcome()))
	urequire.NoError(t, dao.Vote(memberB, p.ID(), commondao.ChoiceYes, "")) // 3*2 >= 2*3
	uassert.Equal(t, int(commondao.OutcomePassed), int(p.ExpectedOutcome()))

}

func TestExecute(cur realm, t *testing.T) {
	t.Run("early passed proposals execute before the deadline", func(t *testing.T) {
		executed := false
		dao := commondao.New(commondao.WithCouncilMember(memberA), commondao.WithProposalKind(probeKind{}))
		p := mustPropose(t, dao, memberA, defaultPropDef{period: time.Hour, executed: &executed})

		urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
		urequire.Equal(t, string(commondao.StatusPassed), string(p.Status()))

		urequire.NoError(t, dao.Execute(p.ID(), cur))
		uassert.True(t, executed)
		uassert.Equal(t, string(commondao.StatusExecuted), string(p.Status()))
		uassert.Equal(t, 0, dao.ActiveProposalsSize())
		uassert.Equal(t, 1, dao.FinishedProposalsSize())
	})

	t.Run("active proposals cannot execute before the deadline", func(t *testing.T) {
		dao := commondao.New(commondao.WithCouncilMember(memberA), commondao.WithProposalKind(probeKind{}))
		p := mustPropose(t, dao, memberA, defaultPropDef{period: time.Hour})

		err := dao.Execute(p.ID(), cur)
		uassert.ErrorIs(t, err, commondao.ErrVotingDeadlineNotMet)
	})

	t.Run("undecided default proposals are dismissed at the deadline", func(t *testing.T) {
		dao := commondao.New(commondao.WithCouncilMember(memberA), commondao.WithProposalKind(probeKind{}))
		p := mustPropose(t, dao, memberA, defaultPropDef{}) // zero voting period

		urequire.NoError(t, dao.Execute(p.ID(), cur))
		uassert.Equal(t, string(commondao.StatusDismissed), string(p.Status()))
		uassert.Equal(t, 1, dao.FinishedProposalsSize())
	})

	t.Run("undecided proposals dismiss before validation", func(t *testing.T) {
		// A proposal that never passed archives as Dismissed even when its
		// definition no longer validates: dismissal is the truthful status
		// and the executor could never have run.
		stale := errors.New("stale")
		dao := commondao.New(commondao.WithCouncilMember(memberA), commondao.WithProposalKind(probeKind{}))
		p := mustPropose(t, dao, memberA, validatingExecPropDef{
			defaultPropDef: defaultPropDef{}, // zero voting period
			err:            &stale,
		})

		urequire.NoError(t, dao.Execute(p.ID(), cur))
		uassert.Equal(t, string(commondao.StatusDismissed), string(p.Status()))
	})

	t.Run("executor errors fail the proposal", func(t *testing.T) {
		dao := commondao.New(commondao.WithCouncilMember(memberA), commondao.WithProposalKind(probeKind{}))
		p := mustPropose(t, dao, memberA, defaultPropDef{period: time.Hour, execErr: errors.New("exec boom")})

		urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
		urequire.NoError(t, dao.Execute(p.ID(), cur))
		uassert.Equal(t, string(commondao.StatusFailed), string(p.Status()))
		uassert.Equal(t, "exec boom", p.StatusReason())
		uassert.Equal(t, 1, dao.FinishedProposalsSize())
	})

	t.Run("emptying council updates fail the proposal", func(t *testing.T) {
		dao := commondao.New(commondao.WithCouncilMember(memberA), commondao.WithProposalKind(probeKind{}))
		p := mustPropose(t, dao, memberA, execFnPropDef{
			propDef: propDef{period: time.Hour},
			fn: func(_ int, sub realm) error {
				return dao.UpdateCouncil(nil, []address{memberA})
			},
		})

		urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
		urequire.NoError(t, dao.Execute(p.ID(), cur))
		uassert.Equal(t, string(commondao.StatusFailed), string(p.Status()))
		uassert.Equal(t, 1, dao.Council().Size(), "expect council unchanged")
	})

	t.Run("early passed proposals still validate", func(t *testing.T) {
		executed := false
		stale := errors.New("stale")
		dao := commondao.New(commondao.WithCouncilMember(memberA), commondao.WithProposalKind(probeKind{}))
		p := mustPropose(t, dao, memberA, validatingExecPropDef{
			defaultPropDef: defaultPropDef{period: time.Hour, executed: &executed},
			err:            &stale,
		})

		urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
		urequire.Equal(t, string(commondao.StatusPassed), string(p.Status()))

		urequire.NoError(t, dao.Execute(p.ID(), cur))
		uassert.Equal(t, string(commondao.StatusFailed), string(p.Status()))
		uassert.Equal(t, "stale", p.StatusReason())
		uassert.False(t, executed, "expect executor to not run when validation fails")
	})

	t.Run("executors cannot re-enter Execute", func(t *testing.T) {
		// The per-DAO executing latch panics on a re-entrant Execute of
		// the same DAO (aborting the tx), even for a different proposal
		// than the one currently executing. Removing the latch would let
		// the re-entrant call proceed.
		var (
			dao   = commondao.New(commondao.WithCouncilMember(memberA), commondao.WithProposalKind(probeKind{}))
			execs int
			pid2  uint64
		)
		p1 := mustPropose(t, dao, memberA, execFnPropDef{
			propDef: propDef{period: time.Hour},
			fn: func(_ int, sub realm) error {
				execs++
				return dao.Execute(pid2, sub) // re-entrant: must panic
			},
		})
		p2 := mustPropose(t, dao, memberA, defaultPropDef{period: time.Hour})
		pid2 = p2.ID()

		urequire.NoError(t, dao.Vote(memberA, p1.ID(), commondao.ChoiceYes, ""))
		urequire.NoError(t, dao.Vote(memberA, pid2, commondao.ChoiceYes, ""))

		uassert.PanicsWithMessage(t, cur, "commondao: re-entrant Execute is not allowed", func() {
			dao.Execute(p1.ID(), cur)
		})
		uassert.Equal(t, 1, execs, "expect executor to run exactly once")
	})

	t.Run("concurrent council updates merge in execution order", func(t *testing.T) {
		dao := commondao.New(
			commondao.WithCouncilMember(memberA),
			commondao.WithCouncilMember(memberB),
			commondao.WithCouncilMember(memberC),
			commondao.WithProposalKind(probeKind{}),
		)
		p1 := mustPropose(t, dao, memberA, execFnPropDef{
			propDef: propDef{period: time.Hour},
			fn: func(_ int, sub realm) error {
				return dao.UpdateCouncil([]address{memberD}, []address{memberA})
			},
		})
		p2 := mustPropose(t, dao, memberB, execFnPropDef{
			propDef: propDef{period: time.Hour},
			fn: func(_ int, sub realm) error {
				// Removing memberA again is an idempotent no-op
				return dao.UpdateCouncil([]address{memberE}, []address{memberA})
			},
		})

		for _, p := range []*commondao.Proposal{p1, p2} {
			urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
			urequire.NoError(t, dao.Vote(memberB, p.ID(), commondao.ChoiceYes, ""))
		}

		urequire.NoError(t, dao.Execute(p1.ID(), cur))
		urequire.NoError(t, dao.Execute(p2.ID(), cur))
		urequire.Equal(t, string(commondao.StatusExecuted), string(p1.Status()))
		urequire.Equal(t, string(commondao.StatusExecuted), string(p2.Status()))

		uassert.Equal(t, 4, dao.Council().Size())
		uassert.False(t, dao.Council().Has(memberA))
		uassert.True(t, dao.Council().Has(memberB))
		uassert.True(t, dao.Council().Has(memberC))
		uassert.True(t, dao.Council().Has(memberD))
		uassert.True(t, dao.Council().Has(memberE))
	})

	t.Run("deleted DAOs reject execution", func(t *testing.T) {
		dao := commondao.New(commondao.WithCouncilMember(memberA), commondao.WithProposalKind(probeKind{}))
		p := mustPropose(t, dao, memberA, propDef{period: time.Hour})
		dao.Dissolve("test")

		err := dao.Execute(p.ID(), cur)
		uassert.ErrorIs(t, err, commondao.ErrDAOIsDeleted)
	})

	t.Run("unknown proposals are rejected", func(t *testing.T) {
		dao := commondao.New(commondao.WithCouncilMember(memberA), commondao.WithProposalKind(probeKind{}))
		err := dao.Execute(404, cur)
		uassert.ErrorIs(t, err, commondao.ErrProposalNotFound)
	})
}

func TestWithdraw(t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithCouncilMember(memberB),
		commondao.WithCouncilMember(memberC),
		commondao.WithProposalKind(probeKind{}),
	)

	// Proposals without votes can be withdrawn
	p := mustPropose(t, dao, memberA, propDef{period: time.Hour})
	urequire.NoError(t, dao.Withdraw(p.ID()))
	uassert.Equal(t, string(commondao.StatusWithdrawn), string(p.Status()))
	uassert.Equal(t, 0, dao.ActiveProposalsSize())
	uassert.Equal(t, 1, dao.FinishedProposalsSize())

	// Withdrawn proposals cannot be withdrawn again
	err := dao.Withdraw(p.ID())
	uassert.ErrorIs(t, err, commondao.ErrProposalNotFound)

	// Proposals with votes cannot be withdrawn
	p = mustPropose(t, dao, memberA, propDef{period: time.Hour})
	urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceAbstain, ""))
	err = dao.Withdraw(p.ID())
	uassert.ErrorIs(t, err, commondao.ErrWithdrawalNotAllowed)

	// Early passed proposals cannot be withdrawn
	p2 := mustPropose(t, dao, memberA, defaultPropDef{period: time.Hour})
	urequire.NoError(t, dao.Vote(memberA, p2.ID(), commondao.ChoiceYes, ""))
	urequire.NoError(t, dao.Vote(memberB, p2.ID(), commondao.ChoiceYes, ""))
	urequire.Equal(t, string(commondao.StatusPassed), string(p2.Status()))
	err = dao.Withdraw(p2.ID())
	uassert.ErrorIs(t, err, commondao.ErrStatusIsNotActive)
}

func TestReadonlyCommonDAO(t *testing.T) {
	parent := commondao.New(commondao.WithName("root"))
	dao := commondao.New(
		commondao.WithID(7),
		commondao.WithName("child"),
		commondao.WithDescription("test"),
		commondao.WithParent(parent),
		commondao.WithCouncilMember(memberA),
		commondao.WithProposalKind(probeKind{}),
	)
	p := mustPropose(t, dao, memberA, propDef{title: "T", body: "B", period: time.Hour})

	view := dao.Readonly()
	uassert.Equal(t, uint64(7), view.ID())
	uassert.Equal(t, "child", view.Name())
	uassert.Equal(t, "test", view.Description())
	uassert.False(t, view.IsDeleted())
	uassert.Equal(t, 1, view.Council().Size())
	uassert.Equal(t, 1, view.ActiveProposalsSize())
	uassert.Equal(t, 0, view.FinishedProposalsSize())
	uassert.Equal(t, 0, view.ChildrenCount())

	parentView, found := view.Parent()
	urequire.True(t, found, "expect parent view")
	uassert.Equal(t, "root", parentView.Name())
	_, found = parentView.Parent()
	uassert.False(t, found)

	pv, found := view.GetProposal(p.ID())
	urequire.True(t, found, "expect proposal view")
	uassert.Equal(t, p.ID(), pv.ID())
	uassert.Equal(t, "T", pv.Title())
	uassert.Equal(t, "B", pv.Body())
	uassert.Equal(t, string(commondao.StatusActive), string(pv.Status()))
	uassert.Equal(t, memberA, pv.Creator())
	uassert.Equal(t, 1, pv.Electorate().Size())
	uassert.Equal(t, 0, pv.VotingRecord().Size())

	visited := 0
	view.IterateActiveProposals(0, 10, false, func(commondao.ReadonlyProposal) bool {
		visited++
		return false
	})
	uassert.Equal(t, 1, visited)

	// WithParent registered dao as a child of parent at construction
	uassert.Equal(t, 1, parentView.ChildrenCount())
	childID := uint64(0)
	parentView.IterateChildren(func(child commondao.ReadonlyCommonDAO) bool {
		childID = child.ID()
		return false
	})
	uassert.Equal(t, uint64(7), childID)

	// A withdrawn proposal is visible through the finished-proposals view
	fp := mustPropose(t, dao, memberA, propDef{period: time.Hour})
	urequire.NoError(t, dao.Withdraw(fp.ID()))
	uassert.Equal(t, 1, view.FinishedProposalsSize())
	finished := 0
	view.IterateFinishedProposals(0, 10, false, func(commondao.ReadonlyProposal) bool {
		finished++
		return false
	})
	uassert.Equal(t, 1, finished)

	// Recorded votes are visible through the voting-record view
	urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
	pv, _ = view.GetProposal(p.ID())
	yes := 0
	pv.VotingRecord().IterateVotesCount(func(c commondao.VoteChoice, count int) bool {
		if c == commondao.ChoiceYes {
			yes = count
		}
		return false
	})
	uassert.Equal(t, 1, yes)

	// Views are live handles
	dao.Dissolve("test")
	uassert.True(t, view.IsDeleted())
}

func TestCharter(t *testing.T) {
	dao := commondao.New(
		commondao.WithName("D"),
		commondao.WithPurpose("to do good"),
		commondao.WithDescription("a test DAO"),
	)
	uassert.Equal(t, "to do good", dao.Purpose())
	uassert.Equal(t, "a test DAO", dao.Description())
	uassert.Equal(t, "to do good", dao.Readonly().Purpose())

	// Purpose defaults empty when unset (the realm enforces non-empty).
	uassert.Equal(t, "", commondao.New().Purpose())
}

func TestAddressAndTreasuryFrozen(t *testing.T) {
	dao := commondao.New(commondao.WithAddress(memberA))
	uassert.Equal(t, string(memberA), string(dao.Address()))

	view := dao.Readonly()
	uassert.Equal(t, string(memberA), string(view.Address()))

	// Unset addresses are empty
	uassert.Equal(t, "", string(commondao.New().Address()))

	uassert.False(t, dao.IsTreasuryFrozen())
	dao.SetTreasuryFrozen(true)
	uassert.True(t, dao.IsTreasuryFrozen())
	uassert.True(t, view.IsTreasuryFrozen())
	dao.SetTreasuryFrozen(false)
	uassert.False(t, view.IsTreasuryFrozen())
}

func TestSetMaxActiveProposals(t *testing.T) {
	dao := commondao.New(commondao.WithCouncilMember(memberA), commondao.WithProposalKind(probeKind{}))

	dao.SetMaxActiveProposals(2)
	_ = mustPropose(t, dao, memberA, propDef{period: time.Hour})
	_ = mustPropose(t, dao, memberA, propDef{period: time.Hour})
	_, err := dao.Propose(memberA, "probe", propDef{period: time.Hour})
	uassert.ErrorIs(t, err, commondao.ErrMaxActiveProposals)

	uassert.Equal(t, 2, dao.MaxActiveProposals())

	// Values below the floor are ignored (the floor is 1, not 0/negative)
	dao.SetMaxActiveProposals(0)
	_, err = dao.Propose(memberA, "probe", propDef{period: time.Hour})
	uassert.ErrorIs(t, err, commondao.ErrMaxActiveProposals)
	dao.SetMaxActiveProposals(-5)
	_, err = dao.Propose(memberA, "probe", propDef{period: time.Hour})
	uassert.ErrorIs(t, err, commondao.ErrMaxActiveProposals)
}

func TestKindRegistry(cur realm, t *testing.T) {
	dao := commondao.New(commondao.WithCouncilMember(memberA))

	// A fresh DAO has no registered kinds
	uassert.False(t, dao.HasKind("probe"))
	uassert.Equal(t, 0, len(dao.KindNames()))

	// Proposing through an unregistered kind is rejected
	_, err := dao.Propose(memberA, "probe", propDef{period: time.Hour})
	uassert.ErrorIs(t, err, commondao.ErrProposalKindNotFound)

	// Registered kinds create proposals
	urequire.NoError(t, dao.RegisterKind(probeKind{}))
	uassert.True(t, dao.HasKind("probe"))
	p := mustPropose(t, dao, memberA, propDef{period: time.Hour})

	// Nil kinds and empty kind names are rejected
	err = dao.RegisterKind(nil)
	uassert.ErrorIs(t, err, commondao.ErrProposalKindRequired)
	err = dao.RegisterKind(namedKind(""))
	uassert.ErrorIs(t, err, commondao.ErrProposalKindRequired)

	// Duplicate kind names are rejected
	err = dao.RegisterKind(probeKind{})
	uassert.ErrorIs(t, err, commondao.ErrProposalKindExists)

	// KindNames is sorted by name
	urequire.NoError(t, dao.RegisterKind(namedKind("zeta")))
	urequire.NoError(t, dao.RegisterKind(namedKind("alpha")))
	names := dao.KindNames()
	urequire.Equal(t, 3, len(names))
	uassert.Equal(t, "alpha", names[0])
	uassert.Equal(t, "probe", names[1])
	uassert.Equal(t, "zeta", names[2])

	// The readonly view mirrors the registry reads
	view := dao.Readonly()
	uassert.True(t, view.HasKind("alpha"))
	uassert.False(t, view.HasKind("beta"))
	uassert.Equal(t, 3, len(view.KindNames()))

	// Deregistering an absent kind is rejected
	err = dao.DeregisterKind("beta")
	uassert.ErrorIs(t, err, commondao.ErrProposalKindNotFound)

	// Deregistering blocks new proposals of the kind...
	urequire.NoError(t, dao.DeregisterKind("probe"))
	uassert.False(t, dao.HasKind("probe"))
	_, err = dao.Propose(memberA, "probe", propDef{period: time.Hour})
	uassert.ErrorIs(t, err, commondao.ErrProposalKindNotFound)

	// ...but never touches in-flight proposals: the frozen definition
	// still votes, decides and executes (vote-integrity)
	urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
	urequire.Equal(t, string(commondao.StatusPassed), string(p.Status()))
	urequire.NoError(t, dao.Execute(p.ID(), cur))
	uassert.Equal(t, string(commondao.StatusExecuted), string(p.Status()))

	// Kind factory errors propagate out of Propose
	boom := errors.New("bad args")
	urequire.NoError(t, dao.RegisterKind(failingKind{err: boom}))
	_, err = dao.Propose(memberA, "failing", nil)
	uassert.ErrorIs(t, err, boom)
	uassert.Equal(t, 0, dao.ActiveProposalsSize(), "expect no proposal from a failing factory")

	// Propose hands the kind factory a readonly view of the host DAO
	var hostID uint64 = 999 // sentinel: New must overwrite it
	urequire.NoError(t, dao.RegisterKind(hostKind{hostID: &hostID}))
	_, err = dao.Propose(memberA, "host", nil)
	urequire.NoError(t, err)
	uassert.Equal(t, dao.ID(), hostID, "expect the kind factory to receive a readonly view of the host DAO")

	// Construction options panic on registration errors
	uassert.PanicsWithMessage(t, cur, commondao.ErrProposalKindRequired.Error(), func() {
		commondao.New(commondao.WithProposalKind(nil))
	})
	uassert.PanicsWithMessage(t, cur, commondao.ErrProposalKindExists.Error(), func() {
		commondao.New(
			commondao.WithProposalKind(probeKind{}),
			commondao.WithProposalKind(probeKind{}),
		)
	})
}
