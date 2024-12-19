package main

// SEND: 200000000ugnot

import governor "gno.land/p/demo/gor/governor"
import "gno.land/r/demo/groups"
import "gno.land/r/demo/users"

func main() {
	println("test governor")
	users.Register("", "gnouser", "my profile")
	evalDAO, groupID := groups.NewGroup("eval_dao")
	println("group created: ", groupID)

	votingPower := governor.NewNoCoinAdapter(evalDAO)
	tally := governor.NewProposalVotes()
	gs := governor.NewGovernorSettings(0, 5, 0, 40, 67, 34)
	_ = governor.NewGovernor("eval_governor", votingPower, tally, gs)
}

// Output:
// test governor
// group created:  1
