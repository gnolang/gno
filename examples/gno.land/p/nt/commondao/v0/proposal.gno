package commondao

import (
	"errors"
	"time"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/bptree/v0"
)

const (
	StatusActive    ProposalStatus = "active"
	StatusPassed    ProposalStatus = "passed"
	StatusDismissed ProposalStatus = "dismissed"
	StatusExecuted  ProposalStatus = "executed"
	StatusFailed    ProposalStatus = "failed"
	StatusWithdrawn ProposalStatus = "withdrawn"
)

const (
	ChoiceNone    VoteChoice = ""
	ChoiceYes     VoteChoice = "YES"
	ChoiceNo      VoteChoice = "NO"
	ChoiceAbstain VoteChoice = "ABSTAIN"
)

// Thresholds for the constitution's default Council voting rules.
const (
	// ThresholdSupermajority passes with "two thirds or more" of the
	// tally denominator. The default for Council decisions.
	ThresholdSupermajority Threshold = iota

	// ThresholdSimpleMajority passes with "more than half" of the
	// tally denominator. The Constitution assigns it to specific
	// decisions, e.g. sub-DAO creation.
	ThresholdSimpleMajority
)

// Outcomes of tallying a proposal under the default Council rules.
const (
	OutcomePending Outcome = iota
	OutcomePassed
	OutcomeDismissed
)

// MaxCustomVoteChoices defines the maximum number of custom
// vote choices that a proposal definition can define.
const MaxCustomVoteChoices = 10

var (
	ErrInvalidCreatorAddress      = errors.New("invalid proposal creator address")
	ErrMaxCustomVoteChoices       = errors.New("max number of custom vote choices exceeded")
	ErrProposalDefinitionRequired = errors.New("proposal definition is required")
	ErrStatusIsNotActive          = errors.New("proposal status is not active")
	ErrCustomChoicesNotAllowed    = errors.New("default tallier definitions cannot customize vote choices")
	ErrMissingTallier             = errors.New("proposal definition must implement DefaultTallier or CustomTallier")
	ErrConflictingTalliers        = errors.New("proposal definition cannot implement both DefaultTallier and CustomTallier")
)

type (
	// ProposalStatus defines a type for different proposal states.
	ProposalStatus string

	// VoteChoice defines a type for proposal vote choices.
	VoteChoice string

	// Threshold defines a type for the default tally thresholds.
	Threshold int

	// Outcome defines a type for default tally outcomes.
	Outcome int

	// ExecFunc defines a type for functions that executes proposals.
	ExecFunc func(realm) error

	// Proposal defines a DAO proposal.
	Proposal struct {
		id             uint64
		status         ProposalStatus
		definition     ProposalDefinition
		creator        address
		record         *VotingRecord
		electorate     *addrset.Set // council snapshot taken at Propose
		statusReason   string
		voteChoices    *bptree.BPTree // string(VoteChoice) -> struct{}
		votingDeadline time.Time
		createdAt      time.Time
	}

	// ProposalDefinition defines an interface for custom proposal definitions.
	// These definitions define proposal content and behavior, essentially
	// allowing the definition of different proposal types.
	//
	// Every definition must also implement exactly one of DefaultTallier
	// (decided by the constitution's default Council voting rules) or
	// CustomTallier (decided by custom logic at the voting deadline).
	ProposalDefinition interface {
		// Title returns the proposal title.
		Title() string

		// Body returns proposal's body.
		// It usually contains description or values that are specific to the proposal,
		// like a description of the proposal's motivation or the list of values that
		// would be applied when the proposal is approved.
		Body() string

		// VotingPeriod returns the period where votes are allowed after proposal creation.
		// It is used to calculate the voting deadline from the proposal's creation date.
		VotingPeriod() time.Duration
	}

	// DefaultTallier defines an interface for proposal definitions that are
	// decided by the constitution's default Council voting rules at the
	// returned threshold. These proposals are evaluated after every recorded
	// vote and can pass or be dismissed before their voting deadline.
	DefaultTallier interface {
		// Threshold returns the tally threshold for passing the proposal.
		Threshold() Threshold
	}

	// CustomTallier defines an interface for proposal definitions that are
	// decided by custom logic. Custom tallies run once, at the voting
	// deadline. The context contains a readonly record with the submitted
	// votes and the proposal's electorate (the council snapshot taken at
	// proposal creation).
	CustomTallier interface {
		// Tally counts the number of votes and verifies if proposal passes.
		Tally(VotingContext) (passes bool, _ error)
	}

	// CapExempt defines an interface for proposal definitions that are not
	// counted against the DAO's active proposals cap. Exempt definitions are
	// instead bounded to one active proposal per creator, so that proposals
	// which remove members (and therefore must never be blockable by a full
	// cap) stay bounded.
	CapExempt interface {
		// CapExempt marks the definition as exempt.
		CapExempt()
	}

	// Validable defines an interface for proposal definitions that require state validation.
	// Validation is done before execution and normally also during proposal rendering.
	Validable interface {
		// Validate validates that the proposal is valid for the current state.
		Validate() error
	}

	// Executable defines an interface for proposal definitions that modify state on approval.
	// Once proposals are executed they are archived and considered finished.
	Executable interface {
		// Executor returns a function to execute the proposal.
		Executor() ExecFunc
	}

	// CustomizableVoteChoices defines an interface for proposal definitions that want
	// to customize the list of allowed voting choices.
	CustomizableVoteChoices interface {
		// CustomVoteChoices returns a list of valid voting choices.
		// Choices are considered valid only when there are at least two possible choices
		// otherwise proposal defaults to using YES, NO and ABSTAIN as valid choices.
		CustomVoteChoices() []VoteChoice
	}
)

// MustValidate validates that a proposal is valid for the current state or panics on error.
func MustValidate(v Validable) {
	if v == nil {
		panic("validable proposal definition is nil")
	}

	if err := v.Validate(); err != nil {
		panic(err)
	}
}

// MustExecute executes an executable proposal or panics on error.
//
// rlm is the realm authority threaded into the executor via cross(rlm).
// MustExecute itself is not a crossing function (rlm sits in a non-first
// parameter slot) because /p/ production code cannot declare crossing
// functions — callers pass `cur` directly as the rlm argument.
func MustExecute(e Executable, rlm realm) {
	if e == nil {
		panic("executable proposal definition is nil")
	}

	fn := e.Executor()
	if fn == nil {
		return
	}

	if err := fn(cross(rlm)); err != nil {
		panic(err)
	}
}

// NewProposal creates a new DAO proposal.
//
// The proposal is created with an empty electorate; only CommonDAO.Propose
// populates the electorate with the council snapshot, so proposals created
// directly through NewProposal reject all votes.
func NewProposal(id uint64, creator address, d ProposalDefinition) (*Proposal, error) {
	if d == nil {
		return nil, ErrProposalDefinitionRequired
	}

	if !creator.IsValid() {
		return nil, ErrInvalidCreatorAddress
	}

	now := time.Now()
	p := &Proposal{
		id:             id,
		status:         StatusActive,
		definition:     d,
		creator:        creator,
		record:         &VotingRecord{},
		electorate:     &addrset.Set{},
		voteChoices:    bptree.NewBPTree32(),
		votingDeadline: now.Add(d.VotingPeriod()),
		createdAt:      now,
	}

	if v, ok := d.(CustomizableVoteChoices); ok {
		// Custom choices make the default YES/NO/ABSTAIN tally unsound
		if _, isDefault := d.(DefaultTallier); isDefault {
			return nil, ErrCustomChoicesNotAllowed
		}

		choices := v.CustomVoteChoices()
		if len(choices) > MaxCustomVoteChoices {
			return nil, ErrMaxCustomVoteChoices
		}

		for _, c := range choices {
			p.voteChoices.Set(string(c), struct{}{})
		}
	}

	// Use default voting choices when the definition returns none or a single vote choice
	if p.voteChoices.Size() < 2 {
		p.voteChoices.Set(string(ChoiceYes), struct{}{})
		p.voteChoices.Set(string(ChoiceNo), struct{}{})
		p.voteChoices.Set(string(ChoiceAbstain), struct{}{})
	}
	return p, nil
}

// ID returns the unique proposal identifier.
func (p Proposal) ID() uint64 {
	return p.id
}

// Definition returns the proposal definition.
// Proposal definitions define proposal content and behavior.
func (p Proposal) Definition() ProposalDefinition {
	return p.definition
}

// Status returns the current proposal status.
func (p Proposal) Status() ProposalStatus {
	return p.status
}

// Creator returns the address of the account that created the proposal.
func (p Proposal) Creator() address {
	return p.creator
}

// CreatedAt returns the time that proposal was created.
func (p Proposal) CreatedAt() time.Time {
	return p.createdAt
}

// VotingRecord returns a record that contains all the votes submitted for the proposal.
func (p Proposal) VotingRecord() *VotingRecord {
	return p.record
}

// Electorate returns the proposal's electorate: a read only view of the
// council snapshot taken when the proposal was created. Members added to
// the council afterwards vote on the next proposal; members removed or
// resigned afterwards remain in the electorate (their silence counts
// against passage).
func (p Proposal) Electorate() *addrset.ReadonlySet {
	return p.electorate.Readonly()
}

// StatusReason returns an optional reason that lead to the current proposal status.
// Reason is mostly useful when a proposal fails.
func (p Proposal) StatusReason() string {
	return p.statusReason
}

// VotingDeadline returns the deadline after which no more votes should be allowed.
func (p Proposal) VotingDeadline() time.Time {
	return p.votingDeadline
}

// VoteChoices returns the list of vote choices allowed for the proposal.
func (p Proposal) VoteChoices() []VoteChoice {
	choices := make([]VoteChoice, 0, p.voteChoices.Size())
	p.voteChoices.Iterate("", "", func(c string, _ any) bool {
		choices = append(choices, VoteChoice(c))
		return false
	})
	return choices
}

// HasVotingDeadlinePassed checks if the voting deadline has been met.
func (p Proposal) HasVotingDeadlinePassed() bool {
	return !time.Now().Before(p.VotingDeadline())
}

// Validate validates that a proposal is valid for the current state.
// Validation is done when the proposal can still be executed (status is
// active or passed) and when the definition supports validation.
func (p Proposal) Validate() error {
	if p.status != StatusActive && p.status != StatusPassed {
		return nil
	}

	if v, ok := p.definition.(Validable); ok {
		return v.Validate()
	}
	return nil
}

// IsVoteChoiceValid checks if a vote choice is valid for the proposal.
func (p Proposal) IsVoteChoiceValid(c VoteChoice) bool {
	return p.voteChoices.Has(string(c))
}

// ExpectedOutcome returns the outcome the proposal would have if it were
// decided with the votes submitted so far. Useful for rendering. Custom
// tally errors report as pending.
func (p Proposal) ExpectedOutcome() Outcome {
	if dt, ok := p.definition.(DefaultTallier); ok {
		return TallyDefault(p.record.Readonly(), p.Electorate(), dt.Threshold())
	}

	if ct, ok := p.definition.(CustomTallier); ok {
		passes, err := ct.Tally(NewVotingContext(p.record.Readonly(), p.Electorate()))
		if err != nil {
			return OutcomePending
		}
		if passes {
			return OutcomePassed
		}
		return OutcomeDismissed
	}
	return OutcomePending
}
