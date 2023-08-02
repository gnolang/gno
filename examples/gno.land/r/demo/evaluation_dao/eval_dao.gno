package evaldao

import (
	"gno.land/p/demo/gor/governor"
	"gno.land/r/demo/groups"
	"std"
)

var (
	evalDAO *groups.Group
	groupID groups.GroupID
	gvr     *governor.Governor
)

func init() {
	evalDAO, groupID = groups.NewGroup("EvalDAO")
	println("group created: ", groupID)
	votingPower := governor.NewNoCoinAdapter(evalDAO)
	tally := governor.NewProposalVotes()
	gs := governor.NewGovernorSettings(0, 5, 0, 40, 67, 34)
	gvr = governor.NewGovernor("eval_governor", votingPower, tally, gs)

	// TODO:
	// render governor
	// render proposal
}

func AddMember(addr std.Address) {
	groups.AddMember(2, "g1vahx7atnv4erxh6lta047h6lta047h6ll85gpy", 55, "metadata3")
}

func TotalVotingPower() uint64 {
	//evalDAO.getNumMembers()
	return 0
}

func GetVotingPowerByAddr(member std.Address) uint64 {
	return uint64(evalDAO.getMemberWeight(member))
}

// TODO:
// render EvalDAO, groupID, num members, member, voting power
// remove member

// propose -> vote -> tally -> render
