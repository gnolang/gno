package admindao

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/seqid"
)

const (
	StatusActive   ProposalStatus = "active"
	StatusPassed                  = "passed"
	StatusFailed                  = "failed"
	StatusExecuted                = "executed"
)

const (
	ChoiceNone VoteChoice = ""
	ChoiceYes             = "yes"
	ChoiceNo              = "no"
)

type (
	ProposalStatus string
	VoteChoice     string

	Vote struct {
		Address std.Address
		Choice  VoteChoice
	}

	Proposal struct {
		id         seqid.ID
		status     ProposalStatus
		definition ProposalDefinition
		creator    std.Address
		votes      *avl.Tree // string(address) -> VoteChoice
		failReason string
		createdAt  time.Time
	}

	ProposalDefinition interface {
		Title() string
		Body() string
		VotingPeriod() time.Duration
		Quorum() float64 // TODO: Quorum should be optional? Default to use majority 20%?
		Validate() error
		Execute() error
	}
)

func (p Proposal) ID() uint64 {
	return uint64(p.id)
}

func (p Proposal) Definition() ProposalDefinition {
	return p.definition
}

func (p Proposal) Status() ProposalStatus {
	return p.status
}

func (p Proposal) Creator() std.Address {
	return p.creator
}

func (p Proposal) CreatedAt() time.Time {
	return p.createdAt
}

func (p Proposal) Votes() rotree.IReadOnlyTree {
	return p.votes
}

func (p Proposal) FailReason() string {
	return p.failReason
}
