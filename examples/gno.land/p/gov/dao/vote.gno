package dao

import (
	"std"
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
	MissingVotes uint64

	ThresholdMet bool
}

func (v VotingStats) GetYayPercent() uint64 {
	// TODO implement
	return 0
}

func (v VotingStats) GetNayPercent() uint64 {
	// TODO implement
	return 0
}

func (v VotingStats) GetAbstainPercent() uint64 {
	// TODO implement
	return 0
}

func (v VotingStats) GetMissingPercent() uint64 {
	// TODO implement
	return 0
}
