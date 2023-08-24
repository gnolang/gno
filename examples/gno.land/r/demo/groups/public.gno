package groups

import (
	"std"

	"gno.land/r/demo/users"
)

//----------------------------------------
// Public facing functions

func GetGroupIDFromName(name string) (GroupID, bool) {
	groupI, exists := gGroupsByName.Get(name)
	if !exists {
		return 0, false
	}
	return groupI.(*Group).id, true
}

func NewGroup(name string) (*Group, GroupID) {
	caller := std.GetOrigCaller()
	usernameOf(caller)
	url := "/r/demo/groups:" + name
	group := newGroup(url, name, caller)
	gidkey := groupIDKey(group.id)
	gGroups.Set(gidkey, group)
	gGroupsByName.Set(name, group)
	return group, group.id
}

func CreateGroup(name string) GroupID {
	std.AssertOriginCall()
	caller := std.GetOrigCaller()
	usernameOf(caller)
	url := "/r/demo/groups:" + name
	group := newGroup(url, name, caller)
	gidkey := groupIDKey(group.id)
	gGroups.Set(gidkey, group)
	gGroupsByName.Set(name, group)
	return group.id
}

func AddMember(gid GroupID, address string, weight int, metadata string) {
	//std.AssertOriginCall()
	caller := std.GetOrigCaller()
	usernameOf(caller)
	group := getGroup(gid)
	if !group.HasPermission(caller, EditPermission) {
		panic("unauthorized to edit group")
	}
	user := users.GetUserByAddress(std.Address(address))
	if user == nil {
		panic("unknown address " + address)
	}
	member := group.newMember(std.Address(address), weight, metadata)
	group.members.Set(address, member)
}

func DeleteGroup(gid GroupID) {
	std.AssertOriginCall()
	caller := std.GetOrigCaller()
	group := getGroup(gid)
	if !group.HasPermission(caller, DeletePermission) {
		panic("unauthorized to delete group")
	}
	group.deleteGroup()
}

func DeleteMember(gid GroupID, addr std.Address) {
	std.AssertOriginCall()
	caller := std.GetOrigCaller()
	group := getGroup(gid)
	if !group.HasPermission(caller, DeletePermission) {
		panic("unauthorized to delete member")
	}
	group.deleteMember(addr)
}

func GetMemberWeight(gid GroupID, addr std.Address) int32 {
	group := getGroup(gid)
	return group.getMemberWeight(addr)
}

func GetMemberWeightAt(gid GroupID, addr std.Address, blockNum int64) int32 {
	group := getGroup(gid)
	return group.getMemberWeightAt(addr, blockNum)
}

func UpdateWeightCheckPoints(gid GroupID, addr std.Address, amount int32, numTiers int32) {
	group := getGroup(gid)
	group.updateWeightSnapshotByTier(addr, amount, numTiers)
}

func GetTotalWeight(gid GroupID) int32 {
	group := getGroup(gid)
	return group.getTotalWeight()
}
