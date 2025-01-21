package admindao

import (
	"std"
	"time"
)

const (
	StatusActive   ProposalStatus = "active"
	StatusPassed                  = "passed"
	StatusFailed                  = "failed"
	StatusExecuted                = "executed"
)

const (
	ChoiceAbstain VoteChoice = ""
	ChoiceYes                = "yes"
	ChoiceNo                 = "no"
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
		record         *VotingRecord
		statusReason   string
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

		// Quorum returns the percentage of members that must vote to be able to pass a proposal.
		// This is an optional value. DAOs use a default value when quorum is zero or invalid.
		// Its value must be between 0 and 1, being 1 = 100% of member votes.
		Quorum() float64

		// Validate validates that the proposal is valid.
		// Validations are optional and allow the validation of the current state before proposal execution.
		Validate() error

		// Execute executes the proposal.
		// Once proposal are executed they are archived and considered finished.
		// Execution allows changing the state after a proposal passes.
		Execute() error
	}
)

// NewProposal creates a new DAO proposal.
func NewProposal(id uint64, creator std.Address, d ProposalDefinition) *Proposal {
	now := time.Now()
	return &Proposal{
		id:             id,
		status:         StatusActive,
		definition:     d,
		creator:        creator,
		record:         &VotingRecord{},
		votingDeadline: now.Add(d.VotingPeriod()),
		createdAt:      now,
	}
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
