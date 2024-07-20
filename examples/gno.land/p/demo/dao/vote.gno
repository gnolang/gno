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

	TotalVotingPower uint64
}

func (v VotingStats) YayPercent() uint64 {
	return (v.YayVotes / v.TotalVotingPower) * 100
}

func (v VotingStats) NayPercent() uint64 {
	return (v.NayVotes / v.TotalVotingPower) * 100
}

func (v VotingStats) AbstainPercent() uint64 {
	return (v.AbstainVotes / v.TotalVotingPower) * 100
}

func (v VotingStats) MissingVotes() uint64 {
	return v.TotalVotingPower - (v.YayVotes + v.NayVotes + v.AbstainVotes)
}

func (v VotingStats) MissingVotesPercent() uint64 {
	return (v.MissingVotes() / v.TotalVotingPower) * 100
}
