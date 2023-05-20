package evaluation

import (
	"std"
)

const (
	VOTE_YES = "YES"
	VOTE_NO  = "NO"
)

type Vote struct {
	voter  std.Address
	option string
}

func NewVote(voter std.Address, option string) Vote {
	v := Vote{
		voter:  voter,
		option: option,
	}
	return v
}
