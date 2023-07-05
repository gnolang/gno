package dao_voting_group

import (
	"std"

	dao_interfaces "gno.land/p/demo/daodao/interfaces"
	"gno.land/p/demo/markdown_utils"
	"gno.land/r/demo/groups"
)

type GRC4Voting struct {
	dao_interfaces.IVotingModule

	groupID groups.GroupID
}

func NewGRC4Voting(groupID groups.GroupID) dao_interfaces.IVotingModule {
	return &GRC4Voting{groupID: groupID}
}

func (v *GRC4Voting) VotingPower(addr std.Address) uint64 {
	return uint64(groups.GetMemberWeightByAddress(v.groupID, addr))
}

func (v *GRC4Voting) TotalPower() uint64 {
	return uint64(groups.GetGroupTotalWeight(v.groupID))
}

func (v *GRC4Voting) Render(path string) string {
	s := "# Group Voting Module\n"
	if groupName, found := groups.GetGroupNameFromID(v.groupID); found {
		s = "# [Group](/r/demo/groups:" + groupName + ") Voting Module\n"
		s += markdown_utils.Indent(groups.Render(groupName))
	} else {
		s += "Group not found"
	}
	return s
}

func (v *GRC4Voting) GetGroupID() groups.GroupID {
	return v.groupID
}
