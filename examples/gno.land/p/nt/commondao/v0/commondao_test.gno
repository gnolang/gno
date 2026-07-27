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
	memberE address = "g1e5tpaw2eaklsgy0dl5rhzkjppe9nlwd8mb4c9a"
	memberF address = "g1fefhhtstt2fhtwnl25gy2mmtwv0go72zvczmgn"
	memberG address = "g1g4ymvyk3axhs9parntwaeq5g02te3f2wc2lahw"
	memberH address = "g1hy6zry03hg5d8le9s7w4fxme6pdrhr55c8lqjs"
	memberI address = "g1manfred47kzduec920z88wfr64ylksmdcedlf5"
)

// propDef is a base custom proposal definition (not a DefaultTallier).
type propDef struct {
	title  string
	body   string
	period time.Duration
	passes bool
	err    error
}

func (d propDef) Title() string               { return d.title }
func (d propDef) Body() string                { return d.body }
func (d propDef) VotingPeriod() time.Duration { return d.period }
func (d propDef) Tally(commondao.VotingContext) (bool, error) {
	return d.passes, d.err
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
	return func(realm) error {
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

// bareDef implements neither DefaultTallier nor CustomTallier.
type bareDef struct{}

func (bareDef) Title() string               { return "Bare" }
func (bareDef) Body() string                { return "" }
func (bareDef) VotingPeriod() time.Duration { return time.Hour }

// dualTallierDef implements both DefaultTallier and CustomTallier.
type dualTallierDef struct{ propDef }

func (dualTallierDef) Threshold() commondao.Threshold { return commondao.ThresholdSupermajority }

// ctxCheckDef records the voting context its custom tally receives.
type ctxCheckDef struct {
	propDef
	votes      *int
	electorate *int
}

func (d ctxCheckDef) Tally(ctx commondao.VotingContext) (bool, error) {
	*d.votes = ctx.VotingRecord.Size()
	*d.electorate = ctx.Electorate.Size()
	return true, nil
}

func TestNew(t *testing.T) {
	dao := commondao.New()
	uassert.Equal(t, uint64(0), dao.ID())
	uassert.Equal(t, "", dao.Name())
	uassert.Equal(t, 0, dao.Council().Size())
	uassert.Equal(t, 0, dao.ActiveProposals().Size())
	uassert.Equal(t, 0, dao.FinishedProposals().Size())
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
	uassert.Equal(t, "root", dao.TopParent().Name())
}

func TestUpdateCouncil(t *testing.T) {
	dao := commondao.New(commondao.WithCouncilMember(memberA))

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
	dao := commondao.New(commondao.WithCouncilMember(memberA))

	p, err := dao.Propose(memberA, propDef{period: time.Hour})
	urequire.NoError(t, err)
	uassert.Equal(t, uint64(1), p.ID())
	uassert.Equal(t, 1, dao.ActiveProposals().Size())
	urequire.True(t, dao.GetProposal(p.ID()) != nil, "expect proposal to be found")

	// The electorate is the council snapshot taken at proposal creation
	uassert.Equal(t, 1, p.Electorate().Size())
	urequire.NoError(t, dao.UpdateCouncil([]address{memberB}, nil))
	uassert.Equal(t, 1, p.Electorate().Size())
	uassert.False(t, p.Electorate().Has(memberB))

	// Definitions must exist and implement exactly one of
	// DefaultTallier/CustomTallier
	_, err = dao.Propose(memberA, nil)
	uassert.ErrorIs(t, err, commondao.ErrProposalDefinitionRequired)
	_, err = dao.Propose(memberA, bareDef{})
	uassert.ErrorIs(t, err, commondao.ErrMissingTallier)
	_, err = dao.Propose(memberA, dualTallierDef{propDef{period: time.Hour}})
	uassert.ErrorIs(t, err, commondao.ErrConflictingTalliers)

	// Deleted DAOs reject proposals
	deleted := commondao.New(commondao.WithCouncilMember(memberA))
	deleted.SetDeleted(true)
	_, err = deleted.Propose(memberA, propDef{})
	uassert.ErrorIs(t, err, commondao.ErrDAOIsDeleted)
}

func TestProposeCap(t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithMaxActiveProposals(2),
	)

	_, err := dao.Propose(memberA, propDef{period: time.Hour})
	urequire.NoError(t, err)
	_, err = dao.Propose(memberA, propDef{period: time.Hour})
	urequire.NoError(t, err)

	// The cap rejects further proposals
	_, err = dao.Propose(memberA, propDef{period: time.Hour})
	uassert.ErrorIs(t, err, commondao.ErrMaxActiveProposals)

	// Cap exempt definitions bypass the cap
	_, err = dao.Propose(memberA, exemptPropDef{defaultPropDef{period: time.Hour}})
	urequire.NoError(t, err)

	// But are bounded to one active exempt proposal per creator
	_, err = dao.Propose(memberA, exemptPropDef{defaultPropDef{period: time.Hour}})
	uassert.ErrorIs(t, err, commondao.ErrMaxCapExemptProposals)
	_, err = dao.Propose(memberB, exemptPropDef{defaultPropDef{period: time.Hour}})
	urequire.NoError(t, err)
}

func TestProposeCapSlotFrees(t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithCouncilMember(memberB),
		commondao.WithCouncilMember(memberC),
		commondao.WithMaxActiveProposals(1),
	)

	p := dao.MustPropose(memberA, defaultPropDef{period: time.Hour})
	_, err := dao.Propose(memberA, defaultPropDef{period: time.Hour})
	urequire.ErrorIs(t, err, commondao.ErrMaxActiveProposals)

	// An immediate NO majority dismissal frees the slot
	urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceNo, ""))
	urequire.NoError(t, dao.Vote(memberB, p.ID(), commondao.ChoiceNo, ""))
	urequire.Equal(t, string(commondao.StatusDismissed), string(p.Status()))

	_, err = dao.Propose(memberA, defaultPropDef{period: time.Hour})
	urequire.NoError(t, err)
}

func TestDissolve(t *testing.T) {
	dao := commondao.New(commondao.WithCouncilMember(memberA))

	// An early passed (but unexecuted) proposal is still in-flight
	p := dao.MustPropose(memberA, defaultPropDef{period: time.Hour})
	urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
	urequire.Equal(t, string(commondao.StatusPassed), string(p.Status()))

	dao.Dissolve("gone")

	uassert.True(t, dao.IsDeleted())
	uassert.Equal(t, 0, dao.ActiveProposals().Size())
	uassert.Equal(t, 1, dao.FinishedProposals().Size())
	uassert.Equal(t, string(commondao.StatusDismissed), string(p.Status()))
	uassert.Equal(t, "gone", p.StatusReason())
}

func TestVote(t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithCouncilMember(memberB),
		commondao.WithCouncilMember(memberC),
	)
	p := dao.MustPropose(memberA, propDef{period: time.Hour})

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

	expired := dao.MustPropose(memberA, propDef{}) // zero voting period
	err = dao.Vote(memberA, expired.ID(), commondao.ChoiceYes, "")
	uassert.ErrorIs(t, err, commondao.ErrVotingDeadlinePassed)

	// Deleted DAOs reject votes
	dao.SetDeleted(true)
	err = dao.Vote(memberA, p.ID(), commondao.ChoiceYes, "")
	uassert.ErrorIs(t, err, commondao.ErrDAOIsDeleted)
}

func TestVoteEarlyTermination(t *testing.T) {
	newDAO := func() *commondao.CommonDAO {
		return commondao.New(
			commondao.WithCouncilMember(memberA),
			commondao.WithCouncilMember(memberB),
			commondao.WithCouncilMember(memberC),
		)
	}

	t.Run("supermajority YES passes immediately", func(t *testing.T) {
		dao := newDAO()
		p := dao.MustPropose(memberA, defaultPropDef{period: time.Hour})

		urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
		uassert.Equal(t, string(commondao.StatusActive), string(p.Status()))

		urequire.NoError(t, dao.Vote(memberB, p.ID(), commondao.ChoiceYes, ""))
		uassert.Equal(t, string(commondao.StatusPassed), string(p.Status()))

		// Passed proposals stay active (pending execution) but reject votes
		uassert.Equal(t, 1, dao.ActiveProposals().Size())
		err := dao.Vote(memberC, p.ID(), commondao.ChoiceNo, "")
		uassert.ErrorIs(t, err, commondao.ErrStatusIsNotActive)
	})

	t.Run("NO majority dismisses immediately", func(t *testing.T) {
		dao := newDAO()
		p := dao.MustPropose(memberA, defaultPropDef{period: time.Hour})

		urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceNo, ""))
		urequire.NoError(t, dao.Vote(memberB, p.ID(), commondao.ChoiceNo, ""))

		uassert.Equal(t, string(commondao.StatusDismissed), string(p.Status()))
		uassert.Equal(t, 0, dao.ActiveProposals().Size())
		uassert.Equal(t, 1, dao.FinishedProposals().Size())
	})

	t.Run("abstain flip can trigger a pass", func(t *testing.T) {
		dao := commondao.New(
			commondao.WithCouncilMember(memberA),
			commondao.WithCouncilMember(memberB),
			commondao.WithCouncilMember(memberC),
			commondao.WithCouncilMember(memberD),
			commondao.WithCouncilMember(memberE),
		)
		p := dao.MustPropose(memberA, defaultPropDef{period: time.Hour})

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
	)

	// Default rule proposals report the live tally
	p := dao.MustPropose(memberA, defaultPropDef{period: time.Hour})
	uassert.Equal(t, int(commondao.OutcomePending), int(p.ExpectedOutcome()))
	urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
	uassert.Equal(t, int(commondao.OutcomePending), int(p.ExpectedOutcome()))
	urequire.NoError(t, dao.Vote(memberB, p.ID(), commondao.ChoiceYes, "")) // 3*2 >= 2*3
	uassert.Equal(t, int(commondao.OutcomePassed), int(p.ExpectedOutcome()))

	// Custom tallier proposals report their Tally result; errors as pending
	p = dao.MustPropose(memberA, propDef{period: time.Hour, passes: true})
	uassert.Equal(t, int(commondao.OutcomePassed), int(p.ExpectedOutcome()))
	p = dao.MustPropose(memberB, propDef{period: time.Hour, passes: false})
	uassert.Equal(t, int(commondao.OutcomeDismissed), int(p.ExpectedOutcome()))
	p = dao.MustPropose(memberC, propDef{period: time.Hour, passes: true, err: errors.New("boom")})
	uassert.Equal(t, int(commondao.OutcomePending), int(p.ExpectedOutcome()))
}

func TestExecute(cur realm, t *testing.T) {
	t.Run("early passed proposals execute before the deadline", func(t *testing.T) {
		executed := false
		dao := commondao.New(commondao.WithCouncilMember(memberA))
		p := dao.MustPropose(memberA, defaultPropDef{period: time.Hour, executed: &executed})

		urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
		urequire.Equal(t, string(commondao.StatusPassed), string(p.Status()))

		urequire.NoError(t, dao.Execute(p.ID(), cur))
		uassert.True(t, executed)
		uassert.Equal(t, string(commondao.StatusExecuted), string(p.Status()))
		uassert.Equal(t, 0, dao.ActiveProposals().Size())
		uassert.Equal(t, 1, dao.FinishedProposals().Size())
	})

	t.Run("active proposals cannot execute before the deadline", func(t *testing.T) {
		dao := commondao.New(commondao.WithCouncilMember(memberA))
		p := dao.MustPropose(memberA, defaultPropDef{period: time.Hour})

		err := dao.Execute(p.ID(), cur)
		uassert.ErrorIs(t, err, commondao.ErrVotingDeadlineNotMet)
	})

	t.Run("undecided default proposals are dismissed at the deadline", func(t *testing.T) {
		dao := commondao.New(commondao.WithCouncilMember(memberA))
		p := dao.MustPropose(memberA, defaultPropDef{}) // zero voting period

		urequire.NoError(t, dao.Execute(p.ID(), cur))
		uassert.Equal(t, string(commondao.StatusDismissed), string(p.Status()))
		uassert.Equal(t, 1, dao.FinishedProposals().Size())
	})

	t.Run("custom definitions tally at the deadline", func(t *testing.T) {
		executed := false
		dao := commondao.New(commondao.WithCouncilMember(memberA))
		p := dao.MustPropose(memberA, execFnPropDef{
			propDef: propDef{passes: true},
			fn: func(realm) error {
				executed = true
				return nil
			},
		})

		urequire.NoError(t, dao.Execute(p.ID(), cur))
		uassert.True(t, executed)
		uassert.Equal(t, string(commondao.StatusExecuted), string(p.Status()))

		p = dao.MustPropose(memberA, propDef{passes: false})
		urequire.NoError(t, dao.Execute(p.ID(), cur))
		uassert.Equal(t, string(commondao.StatusDismissed), string(p.Status()))

		p = dao.MustPropose(memberA, propDef{passes: true, err: errors.New("boom")})
		urequire.NoError(t, dao.Execute(p.ID(), cur))
		uassert.Equal(t, string(commondao.StatusFailed), string(p.Status()))
		uassert.Equal(t, "boom", p.StatusReason())
	})

	t.Run("executor errors fail the proposal", func(t *testing.T) {
		dao := commondao.New(commondao.WithCouncilMember(memberA))
		p := dao.MustPropose(memberA, defaultPropDef{period: time.Hour, execErr: errors.New("exec boom")})

		urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
		urequire.NoError(t, dao.Execute(p.ID(), cur))
		uassert.Equal(t, string(commondao.StatusFailed), string(p.Status()))
		uassert.Equal(t, "exec boom", p.StatusReason())
		uassert.Equal(t, 1, dao.FinishedProposals().Size())
	})

	t.Run("emptying council updates fail the proposal", func(t *testing.T) {
		dao := commondao.New(commondao.WithCouncilMember(memberA))
		p := dao.MustPropose(memberA, execFnPropDef{
			propDef: propDef{passes: true},
			fn: func(realm) error {
				return dao.UpdateCouncil(nil, []address{memberA})
			},
		})

		urequire.NoError(t, dao.Execute(p.ID(), cur))
		uassert.Equal(t, string(commondao.StatusFailed), string(p.Status()))
		uassert.Equal(t, 1, dao.Council().Size(), "expect council unchanged")
	})

	t.Run("early passed proposals still validate", func(t *testing.T) {
		executed := false
		stale := errors.New("stale")
		dao := commondao.New(commondao.WithCouncilMember(memberA))
		p := dao.MustPropose(memberA, validatingExecPropDef{
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

	t.Run("executors cannot re-enter Execute for the same proposal", func(t *testing.T) {
		var (
			dao      = commondao.New(commondao.WithCouncilMember(memberA))
			execs    int
			innerErr error
			pid      uint64
		)
		p := dao.MustPropose(memberA, execFnPropDef{
			propDef: propDef{passes: true},
			fn: func(cur realm) error {
				execs++
				innerErr = dao.Execute(pid, cur)
				return nil
			},
		})
		pid = p.ID()

		urequire.NoError(t, dao.Execute(pid, cur))
		uassert.Equal(t, 1, execs, "expect executor to run exactly once")
		uassert.ErrorIs(t, innerErr, commondao.ErrProposalNotFound)
		uassert.Equal(t, string(commondao.StatusExecuted), string(p.Status()))
		uassert.Equal(t, 1, dao.FinishedProposals().Size())
	})

	t.Run("custom talliers receive the record and electorate", func(t *testing.T) {
		dao := commondao.New(
			commondao.WithCouncilMember(memberA),
			commondao.WithCouncilMember(memberB),
		)
		var votes, electorate int
		p := dao.MustPropose(memberA, ctxCheckDef{
			propDef:    propDef{period: time.Hour},
			votes:      &votes,
			electorate: &electorate,
		})
		urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))

		testing.SkipHeights(721) // 5s blocks: pass the 1h voting deadline
		urequire.NoError(t, dao.Execute(p.ID(), cur))
		uassert.Equal(t, string(commondao.StatusExecuted), string(p.Status()))
		uassert.Equal(t, 1, votes)
		uassert.Equal(t, 2, electorate)
	})

	t.Run("concurrent council updates merge in execution order", func(t *testing.T) {
		dao := commondao.New(
			commondao.WithCouncilMember(memberA),
			commondao.WithCouncilMember(memberB),
			commondao.WithCouncilMember(memberC),
		)
		p1 := dao.MustPropose(memberA, execFnPropDef{
			propDef: propDef{passes: true},
			fn: func(realm) error {
				return dao.UpdateCouncil([]address{memberD}, []address{memberA})
			},
		})
		p2 := dao.MustPropose(memberB, execFnPropDef{
			propDef: propDef{passes: true},
			fn: func(realm) error {
				// Removing memberA again is an idempotent no-op
				return dao.UpdateCouncil([]address{memberE}, []address{memberA})
			},
		})

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
		dao := commondao.New(commondao.WithCouncilMember(memberA))
		p := dao.MustPropose(memberA, propDef{passes: true})
		dao.SetDeleted(true)

		err := dao.Execute(p.ID(), cur)
		uassert.ErrorIs(t, err, commondao.ErrDAOIsDeleted)
	})

	t.Run("unknown proposals are rejected", func(t *testing.T) {
		dao := commondao.New(commondao.WithCouncilMember(memberA))
		err := dao.Execute(404, cur)
		uassert.ErrorIs(t, err, commondao.ErrProposalNotFound)
	})
}

func TestWithdraw(t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithCouncilMember(memberB),
		commondao.WithCouncilMember(memberC),
	)

	// Proposals without votes can be withdrawn
	p := dao.MustPropose(memberA, propDef{period: time.Hour})
	urequire.NoError(t, dao.Withdraw(p.ID()))
	uassert.Equal(t, string(commondao.StatusWithdrawn), string(p.Status()))
	uassert.Equal(t, 0, dao.ActiveProposals().Size())
	uassert.Equal(t, 1, dao.FinishedProposals().Size())

	// Withdrawn proposals cannot be withdrawn again
	err := dao.Withdraw(p.ID())
	uassert.ErrorIs(t, err, commondao.ErrProposalNotFound)

	// Proposals with votes cannot be withdrawn
	p = dao.MustPropose(memberA, propDef{period: time.Hour})
	urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceAbstain, ""))
	err = dao.Withdraw(p.ID())
	uassert.ErrorIs(t, err, commondao.ErrWithdrawalNotAllowed)

	// Early passed proposals cannot be withdrawn
	p2 := dao.MustPropose(memberA, defaultPropDef{period: time.Hour})
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
	)
	p := dao.MustPropose(memberA, propDef{title: "T", body: "B", period: time.Hour})

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
	uassert.Equal(t, 3, len(pv.VoteChoices()))

	visited := 0
	view.IterateActiveProposals(0, 10, false, func(commondao.ReadonlyProposal) bool {
		visited++
		return false
	})
	uassert.Equal(t, 1, visited)

	// Registering dao as a child of parent is visible through the view
	parent.Children().Append(dao)
	uassert.Equal(t, 1, parentView.ChildrenCount())
	childID := uint64(0)
	parentView.IterateChildren(func(child commondao.ReadonlyCommonDAO) bool {
		childID = child.ID()
		return false
	})
	uassert.Equal(t, uint64(7), childID)

	// A withdrawn proposal is visible through the finished-proposals view
	fp := dao.MustPropose(memberA, propDef{period: time.Hour})
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
	dao.SetDeleted(true)
	uassert.True(t, view.IsDeleted())
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
	dao := commondao.New(commondao.WithCouncilMember(memberA))

	dao.SetMaxActiveProposals(2)
	_ = dao.MustPropose(memberA, propDef{period: time.Hour})
	_ = dao.MustPropose(memberA, propDef{period: time.Hour})
	_, err := dao.Propose(memberA, propDef{period: time.Hour})
	uassert.ErrorIs(t, err, commondao.ErrMaxActiveProposals)

	// Values below the floor are ignored (the floor is 1, not 0/negative)
	dao.SetMaxActiveProposals(0)
	_, err = dao.Propose(memberA, propDef{period: time.Hour})
	uassert.ErrorIs(t, err, commondao.ErrMaxActiveProposals)
	dao.SetMaxActiveProposals(-5)
	_, err = dao.Propose(memberA, propDef{period: time.Hour})
	uassert.ErrorIs(t, err, commondao.ErrMaxActiveProposals)
}
