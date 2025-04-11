package evaluation

import "chain"

const (
	VoteYes = "YES"
	VoteNo  = "NO"
)

type Vote struct {
	voter  chain.Address
	option string
}

func NewVote(voter chain.Address, option string) Vote {
	v := Vote{
		voter:  voter,
		option: option,
	}
	return v
}
