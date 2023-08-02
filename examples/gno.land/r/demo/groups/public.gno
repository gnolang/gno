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

func AddMember(gid GroupID, address string, weight int, metadata string) MemberID {
	std.AssertOriginCall()
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
	mid := group.lastMemberID
	member := group.newMember(mid, std.Address(address), weight, metadata)
	midkey := memberIDKey(mid)
	group.members.Set(midkey, member)
	mid++
	group.lastMemberID = mid
	return member.id
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

func DeleteMember(gid GroupID, mid MemberID) {
	std.AssertOriginCall()
	caller := std.GetOrigCaller()
	group := getGroup(gid)
	if !group.HasPermission(caller, DeletePermission) {
		panic("unauthorized to delete member")
	}
	group.deleteMember(mid)
}

func GetMember(gid GroupID, mid MemberID) {
	caller := std.GetOrigCaller()
	group := getGroup(gid)
	if !group.HasPermission(caller, DeletePermission) {
		panic("unauthorized to delete member")
	}
	group.deleteMember(mid)
}
