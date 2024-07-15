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
