package dao

import (
	"std"
)

const (
	VoteAddedEvent = "VoteAdded" // emitted when a vote was cast for a proposal

	VoteAddedIDKey     = "proposal-id"
	VoteAddedAuthorKey = "author"
	VoteAddedOptionKey = "option"
)

type VoteOption string

const (
	YesVote     VoteOption = "YES"
	NoVote      VoteOption = "NO"
	AbstainVote VoteOption = "ABSTAIN"
)

func (v VoteOption) String() string {
	return string(v)
}

// Vote encompasses a single GOVDAO vote
type Vote struct {
	Address std.Address // the address of the voter
	Option  VoteOption  // the voting option
}

// Stats encompasses the proposal voting stats
type Stats struct {
	YayVotes     uint64
	NayVotes     uint64
	AbstainVotes uint64

	TotalVotingPower uint64
}

// YayPercent returns the percentage (0-100) of the yay votes
// in relation to the total voting power
func (v Stats) YayPercent() uint64 {
	return v.YayVotes * 100 / v.TotalVotingPower
}

// NayPercent returns the percentage (0-100) of the nay votes
// in relation to the total voting power
func (v Stats) NayPercent() uint64 {
	return v.NayVotes * 100 / v.TotalVotingPower
}

// AbstainPercent returns the percentage (0-100) of the abstain votes
// in relation to the total voting power
func (v Stats) AbstainPercent() uint64 {
	return v.AbstainVotes * 100 / v.TotalVotingPower
}

// MissingVotes returns the summed voting power that has not
// participated in proposal voting yet
func (v Stats) MissingVotes() uint64 {
	return v.TotalVotingPower - (v.YayVotes + v.NayVotes + v.AbstainVotes)
}

// MissingVotesPercent returns the percentage (0-100) of the missing votes
// in relation to the total voting power
func (v Stats) MissingVotesPercent() uint64 {
	return v.MissingVotes() * 100 / v.TotalVotingPower
}
