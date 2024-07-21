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

// VotingStats encompasses the proposal voting stats
type VotingStats struct {
	YayVotes     uint64
	NayVotes     uint64
	AbstainVotes uint64

	TotalVotingPower uint64
}

// YayPercent returns the percentage (0-100) of the yay votes
// in relation to the total voting power
func (v VotingStats) YayPercent() uint64 {
	return (v.YayVotes / v.TotalVotingPower) * 100
}

// NayPercent returns the percentage (0-100) of the nay votes
// in relation to the total voting power
func (v VotingStats) NayPercent() uint64 {
	return (v.NayVotes / v.TotalVotingPower) * 100
}

// AbstainPercent returns the percentage (0-100) of the abstain votes
// in relation to the total voting power
func (v VotingStats) AbstainPercent() uint64 {
	return (v.AbstainVotes / v.TotalVotingPower) * 100
}

// MissingVotes returns the summed voting power that has not
// participated in proposal voting yet
func (v VotingStats) MissingVotes() uint64 {
	return v.TotalVotingPower - (v.YayVotes + v.NayVotes + v.AbstainVotes)
}

// MissingVotesPercent returns the percentage (0-100) of the missing votes
// in relation to the total voting power
func (v VotingStats) MissingVotesPercent() uint64 {
	return (v.MissingVotes() / v.TotalVotingPower) * 100
}
