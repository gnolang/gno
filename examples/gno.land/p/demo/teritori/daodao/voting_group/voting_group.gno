package dao_voting_group

import (
	"std"

	dao_interfaces "gno.land/p/demo/teritori/daodao/interfaces"
	"gno.land/p/demo/markdown_utils"
	"gno.land/r/demo/groups"
)

type VotingGroup struct {
	dao_interfaces.IVotingModule

	groupID groups.GroupID
}

func NewVotingGroup(groupID groups.GroupID) dao_interfaces.IVotingModule {
	return &VotingGroup{groupID: groupID}
}

func (v *VotingGroup) VotingPowerAtHeight(addr std.Address, height int64) uint64 {
	return uint64(groups.GetMemberWeightByAddress(v.groupID, addr, height))
}

func (v *VotingGroup) TotalPowerAtHeight(height int64) uint64 {
	return uint64(groups.GetGroupTotalWeight(v.groupID, height))
}

func (v *VotingGroup) Render(path string) string {
	s := "# Group Voting Module\n"
	if groupName, found := groups.GetGroupNameFromID(v.groupID); found {
		s = "# [Group](/r/demo/groups:" + groupName + ") Voting Module\n"
		s += markdown_utils.Indent(groups.Render(groupName))
	} else {
		s += "Group not found"
	}
	return s
}

func (v *VotingGroup) GetGroupID() groups.GroupID {
	return v.groupID
}
