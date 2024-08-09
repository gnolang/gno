package govdao

import (
	"std"

	"gno.land/p/demo/ufmt"
)

const (
	yay = "YES"
	nay = "NO"

	msgNoMoreVotesAllowed = "no more votes allowed"
	msgAlreadyVoted       = "caller already voted"
	msgWrongVotingValue   = "voting values must be YES or NO"
)

func NewPercentageVoter(percent int) *PercentageVoter {
	if percent < 0 || percent > 100 {
		panic("percent value must be between 0 and 100")
	}

	return &PercentageVoter{
		percentage: percent,
	}
}

// PercentageVoter is a system based on the amount of received votes.
// When the specified threshold is reached, the voting process finishes.
type PercentageVoter struct {
	percentage int

	voters []std.Address
	yes    int
	no     int
}

var _ Voter = (*PercentageVoter)(nil)

func (pv *PercentageVoter) IsAccepted(voters []std.Address) bool {
	if len(voters) == 0 {
		return true // special case
	}

	return pv.percent(voters) >= pv.percentage
}

func (pv *PercentageVoter) IsFinished(voters []std.Address) bool {
	return pv.yes+pv.no >= len(voters)
}

func (pv *PercentageVoter) Status(voters []std.Address) string {
	return ufmt.Sprintf("- YAY: %d\n\n- NAY: %d\n\n- Percent: %d\n\n- Members: %d", pv.yes, pv.no, pv.percent(voters), len(voters))
}

func (pv *PercentageVoter) Vote(voters []std.Address, caller std.Address, flag string) {
	if pv.IsFinished(voters) {
		panic(msgNoMoreVotesAllowed)
	}

	if pv.alreadyVoted(caller) {
		panic(msgAlreadyVoted)
	}

	switch flag {
	case yay:
		pv.yes++
		pv.voters = append(pv.voters, caller)
	case nay:
		pv.no++
		pv.voters = append(pv.voters, caller)
	default:
		panic(msgWrongVotingValue)
	}
}

func (pv *PercentageVoter) percent(voters []std.Address) int {
	if len(voters) == 0 {
		return 0
	}

	return int((float32(pv.yes) / float32(len(voters))) * 100)
}

func (pv *PercentageVoter) alreadyVoted(addr std.Address) bool {
	for _, v := range pv.voters {
		if v == addr {
			return true
		}
	}

	return false
}
