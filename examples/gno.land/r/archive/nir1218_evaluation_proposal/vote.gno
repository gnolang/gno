package evaluation

const (
	VoteYes = "YES"
	VoteNo  = "NO"
)

type Vote struct {
	voter  address
	option string
}

func NewVote(voter address, option string) Vote {
	v := Vote{
		voter:  voter,
		option: option,
	}
	return v
}
