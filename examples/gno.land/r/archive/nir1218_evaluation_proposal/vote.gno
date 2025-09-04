package evaluation

import "std"

const (
	VoteYes = "YES"
	VoteNo  = "NO"
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
