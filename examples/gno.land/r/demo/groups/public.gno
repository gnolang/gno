package groups

import (
	"std"

	"gno.land/r/sys/users"
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

func CreateGroup(cur realm, name string) GroupID {
	std.AssertOriginCall()
	caller := std.OriginCaller()
	usernameOf(caller)
	url := "/r/demo/groups:" + name
	group := newGroup(url, name, caller)
	gidkey := groupIDKey(group.id)
	gGroups.Set(gidkey, group)
	gGroupsByName.Set(name, group)
	return group.id
}

func AddMember(cur realm, gid GroupID, address_XXX string, weight int, metadata string) MemberID {
	std.AssertOriginCall()
	caller := std.OriginCaller()
	usernameOf(caller)
	group := getGroup(gid)
	if !group.HasPermission(caller, EditPermission) {
		panic("unauthorized to edit group")
	}
	user := users.ResolveAddress(std.Address(address_XXX))
	if user == nil {
		panic("unknown address " + address_XXX)
	}
	mid := group.lastMemberID
	member := group.newMember(mid, std.Address(address_XXX), weight, metadata)
	midkey := memberIDKey(mid)
	group.members.Set(midkey, member)
	mid++
	group.lastMemberID = mid
	return member.id
}

func DeleteGroup(cur realm, gid GroupID) {
	std.AssertOriginCall()
	caller := std.OriginCaller()
	group := getGroup(gid)
	if !group.HasPermission(caller, DeletePermission) {
		panic("unauthorized to delete group")
	}
	group.deleteGroup()
}

func DeleteMember(cur realm, gid GroupID, mid MemberID) {
	std.AssertOriginCall()
	caller := std.OriginCaller()
	group := getGroup(gid)
	if !group.HasPermission(caller, DeletePermission) {
		panic("unauthorized to delete member")
	}
	group.deleteMember(mid)
}
