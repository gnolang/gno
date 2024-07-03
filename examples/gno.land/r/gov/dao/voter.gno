package govdao

import (
	"std"
)

const YAY = "YES"
const NAY = "NO"

func NewPercentageVoter(percent int) *PercentageVoter {
	if percent < 0 || percent > 100 {
		panic("percent value must be between 0 and 100")
	}

	return &PercentageVoter{
		percentage: percent,
	}
}

// PercentageVoter is a system based on the amount of received votes.
// When the specified treshold is reached, the voting process finishes.
type PercentageVoter struct {
	percentage int

	yes int
	no  int
}

func (pv *PercentageVoter) Accepted(voters []std.Address) bool {
	if len(voters) == 0 {
		return true // special case
	}

	return (pv.yes/len(voters))*100 >= pv.percentage
}

func (pv *PercentageVoter) Finished(voters []std.Address) bool {
	return pv.yes+pv.no >= len(voters)
}

func (pv *PercentageVoter) Vote(voters []std.Address, flag string) {
	if (pv.no + pv.yes) == len(voters) {
		panic("no more votes allowed")
	}

	if flag == NAY {
		pv.no++
		return
	}

	if flag == YAY {
		pv.yes++
		return
	}

	panic("voting values must be YES or NO")
}
