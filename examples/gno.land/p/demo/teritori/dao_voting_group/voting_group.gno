package dao_voting_group

import (
	"std"

	dao_interfaces "gno.land/p/demo/teritori/dao_interfaces"
	"gno.land/p/demo/teritori/markdown_utils"
	"gno.land/p/demo/teritori/ujson"
	"gno.land/r/demo/teritori/groups"
)

type VotingGroup struct {
	dao_interfaces.IVotingModule

	groupID groups.GroupID
}

func NewVotingGroup(groupID groups.GroupID) dao_interfaces.IVotingModule {
	return &VotingGroup{groupID: groupID}
}

func (v *VotingGroup) Info() dao_interfaces.ModuleInfo {
	return dao_interfaces.ModuleInfo{
		Kind:    "GroupVoting",
		Version: "0.1.0",
	}
}

func (v *VotingGroup) ConfigJSON() string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "groupId", Value: v.groupID},
	})
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
