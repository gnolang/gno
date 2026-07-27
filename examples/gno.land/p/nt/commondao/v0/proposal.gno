package commondao

import (
	"errors"
	"time"

	"gno.land/p/moul/addrset"
)

const (
	StatusActive    ProposalStatus = "active"
	StatusPassed    ProposalStatus = "passed"
	StatusDismissed ProposalStatus = "dismissed"
	StatusExecuted  ProposalStatus = "executed"
	StatusFailed    ProposalStatus = "failed"
	StatusWithdrawn ProposalStatus = "withdrawn"
)

// Vote choices, fixed by the Common DAO Spec's default voting rules.
const (
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

var (
	ErrInvalidCreatorAddress      = errors.New("invalid proposal creator address")
	ErrInvalidVoterAddress        = errors.New("invalid voter address")
	ErrProposalDefinitionRequired = errors.New("proposal definition is required")
	ErrStatusIsNotActive          = errors.New("proposal status is not active")
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
		votingDeadline time.Time
		createdAt      time.Time
	}

	// ProposalDefinition defines an interface for custom proposal definitions.
	// These definitions define proposal content and behavior, essentially
	// allowing the definition of different proposal types.
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

		// Threshold returns the tally threshold for passing the proposal.
		// Proposals are decided by the constitution's default Council voting
		// rules: re-evaluated after every recorded vote, they can pass or be
		// dismissed before their voting deadline.
		Threshold() Threshold
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
)

// newProposal creates a new DAO proposal.
//
// The proposal is created with an empty electorate; Propose populates it
// with the council snapshot.
func newProposal(id uint64, creator address, d ProposalDefinition) (*Proposal, error) {
	if !creator.IsValid() {
		return nil, ErrInvalidCreatorAddress
	}

	now := time.Now()
	return &Proposal{
		id:             id,
		status:         StatusActive,
		definition:     d,
		creator:        creator,
		record:         &VotingRecord{},
		electorate:     &addrset.Set{},
		votingDeadline: now.Add(d.VotingPeriod()),
		createdAt:      now,
	}, nil
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

// VotingRecord returns a read only record with the votes submitted for
// the proposal. Votes are recorded through CommonDAO.Vote only.
func (p Proposal) VotingRecord() ReadonlyVotingRecord {
	return p.record.Readonly()
}

// Electorate returns the proposal's electorate: a read only view of the
// council snapshot taken when the proposal was created. Members added to
// the council afterwards vote on the next proposal; members removed or
// resigned afterwards remain in the electorate (their silence counts
// against passage).
func (p Proposal) Electorate() *addrset.ReadonlySet {
	return p.electorate.Readonly()
}

// StatusReason returns an optional reason that led to the current proposal status.
// Reason is mostly useful when a proposal fails.
func (p Proposal) StatusReason() string {
	return p.statusReason
}

// VotingDeadline returns the deadline after which no more votes should be allowed.
func (p Proposal) VotingDeadline() time.Time {
	return p.votingDeadline
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

// ExpectedOutcome returns the outcome the proposal would have if it were
// decided with the votes submitted so far. Useful for rendering.
func (p Proposal) ExpectedOutcome() Outcome {
	return TallyDefault(p.record.Readonly(), p.Electorate(), p.definition.Threshold())
}
