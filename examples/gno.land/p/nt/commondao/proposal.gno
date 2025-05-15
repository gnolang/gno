package commondao

import (
	"errors"
	"std"
	"time"

	"gno.land/p/demo/avl"
)

const (
	StatusActive ProposalStatus = "active"
	StatusFailed                = "failed"
	StatusPassed                = "passed"
)

const (
	ChoiceNone       VoteChoice = ""
	ChoiceYes                   = "YES"
	ChoiceNo                    = "NO"
	ChoiceNoWithVeto            = "NO WITH VETO"
	ChoiceAbstain               = "ABSTAIN"
)

const (
	QuorumOneThird     float64 = 0.33 // percentage
	QuorumHalf                 = 0.5
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

	// Proposal defines a DAO proposal.
	Proposal struct {
		id             uint64
		status         ProposalStatus
		definition     ProposalDefinition
		creator        std.Address
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

		// Body returns the proposal body.
		// It usually contains the proposal description and other elements like proposal parameters.
		Body() string

		// VotingPeriod returns the period where votes are allowed after proposal creation.
		// No more votes should be allowed once this period is met. It is used to calculate
		// the voting deadline from the proposal's creationd date.
		VotingPeriod() time.Duration

		// Tally counts the number of votes and verifies if proposal passes.
		Tally(ReadonlyVotingRecord, MemberSet) (passes bool, _ error)
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
		// Execute executes the proposal.
		// This method must allow crossing.
		Execute() error
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

// NewProposal creates a new DAO proposal.
func NewProposal(id uint64, creator std.Address, d ProposalDefinition) (*Proposal, error) {
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
func (p Proposal) Creator() std.Address {
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

// IsQuorumReached checks if a participation quorum is reach.
func IsQuorumReached(quorum float64, r ReadonlyVotingRecord, members MemberSet) bool {
	if members.Size() <= 0 || quorum <= 0 {
		return false
	}

	var votesCount int
	for _, c := range r.Choices() {
		// Don't count explicit abstentions or invalid votes
		if c == ChoiceNone || c == ChoiceAbstain {
			continue
		}

		votesCount += r.VoteCount(c)
	}

	percentage := float64(votesCount) / float64(members.Size())
	return percentage >= quorum
}
