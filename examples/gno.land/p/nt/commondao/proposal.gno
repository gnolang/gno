package commondao

import (
	"errors"
	"time"

	"gno.land/p/nt/avl"
)

const (
	StatusActive    ProposalStatus = "active"
	StatusPassed                   = "passed"
	StatusRejected                 = "rejected"
	StatusExecuted                 = "executed"
	StatusFailed                   = "failed"
	StatusWithdrawn                = "withdrawn"
)

const (
	ChoiceNone       VoteChoice = ""
	ChoiceYes                   = "YES"
	ChoiceNo                    = "NO"
	ChoiceNoWithVeto            = "NO WITH VETO"
	ChoiceAbstain               = "ABSTAIN"
)

const (
	QuorumOneThird     float64 = 0.33 // percentage, checked as >= than quorum
	QuorumMoreThanHalf         = 0.51
	QuorumTwoThirds            = 0.66
	QuorumThreeFourths         = 0.75
	QuorumFull                 = 1
)

// MaxCustomVoteChoices defines the maximum number of custom
// vote choices that a proposal definition can define.
const MaxCustomVoteChoices = 10

var (
	ErrInvalidCreatorAddress      = errors.New("invalid proposal creator address")
	ErrMaxCustomVoteChoices       = errors.New("max number of custom vote choices exceeded")
	ErrProposalDefinitionRequired = errors.New("proposal definition is required")
	ErrNoQuorum                   = errors.New("no quorum")
	ErrStatusIsNotActive          = errors.New("proposal status is not active")
)

type (
	// ProposalStatus defines a type for different proposal states.
	ProposalStatus string

	// VoteChoice defines a type for proposal vote choices.
	VoteChoice string

	// ExecFunc defines a type for functions that executes proposals.
	ExecFunc func(realm) error

	// Proposal defines a DAO proposal.
	Proposal struct {
		id             uint64
		status         ProposalStatus
		definition     ProposalDefinition
		creator        address
		record         *VotingRecord // TODO: Add support for multiple voting records
		statusReason   string
		voteChoices    *avl.Tree // string(VoteChoice) -> struct{}
		votingDeadline time.Time
		createdAt      time.Time
	}

	// ProposalDefinition defines an interface for custom proposal definitions.
	// These definitions define proposal content and behavior, they esentially
	// allow the definition for different proposal types.
	ProposalDefinition interface {
		// Title returns the proposal title.
		Title() string

		// Body returns proposal's body.
		// It usually contains description or values that are specific to the proposal,
		// like a description of the proposal's motivation or the list of values that
		// would be applied when the proposal is approved.
		Body() string

		// VotingPeriod returns the period where votes are allowed after proposal creation.
		// It is used to calculate the voting deadline from the proposal's creationd date.
		VotingPeriod() time.Duration

		// Tally counts the number of votes and verifies if proposal passes.
		// It receives a voting context containing a readonly record with the votes
		// that has been submitted for the proposal and also the list of DAO members.
		Tally(VotingContext) (passes bool, _ error)
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
func MustExecute(e Executable) {
	if e == nil {
		panic("executable proposal definition is nil")
	}

	fn := e.Executor()
	if fn == nil {
		return
	}

	if err := fn(cross); err != nil {
		panic(err)
	}
}

// NewProposal creates a new DAO proposal.
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
		voteChoices:    avl.NewTree(),
		votingDeadline: now.Add(d.VotingPeriod()),
		createdAt:      now,
	}

	if v, ok := d.(CustomizableVoteChoices); ok {
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

// ID returns the unique proposal identifies.
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

// StatusReason returns an optional reason that lead to the current proposal status.
// Reason is mostyl useful when a proposal fails.
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
// Validation is done when proposal status is active and when the definition supports validation.
func (p Proposal) Validate() error {
	if p.status != StatusActive {
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

// Tally counts votes and updates proposal status with the current outcome.
// Proposal status is updated to "passed" when proposal is approved
// or to "rejected" if proposal doesn't pass.
func (p *Proposal) Tally(members MemberStorage) error {
	if p.status != StatusActive {
		return ErrStatusIsNotActive
	}

	ctx := MustNewVotingContext(p.VotingRecord(), members)
	passes, err := p.Definition().Tally(ctx)
	if err != nil {
		return err
	}

	if passes {
		p.status = StatusPassed
	} else {
		p.status = StatusRejected
	}
	return nil
}

// IsQuorumReached checks if a participation quorum is reach.
func IsQuorumReached(quorum float64, r ReadonlyVotingRecord, members ReadonlyMemberStorage) bool {
	if members.Size() <= 0 || quorum <= 0 {
		return false
	}

	var totalCount int
	r.IterateVotesCount(func(c VoteChoice, voteCount int) bool {
		// Don't count explicit abstentions or invalid votes
		if c != ChoiceNone && c != ChoiceAbstain {
			totalCount += r.VoteCount(c)
		}
		return false
	})

	percentage := float64(totalCount) / float64(members.Size())
	return percentage >= quorum
}
