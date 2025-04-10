package groups

import (
	"chain"
	"chain/runtime"

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

func CreateGroup(name string) GroupID {
	runtime.AssertOriginCall()
	caller := runtime.OriginCaller()
	usernameOf(caller)
	url := "/r/demo/groups:" + name
	group := newGroup(url, name, caller)
	gidkey := groupIDKey(group.id)
	gGroups.Set(gidkey, group)
	gGroupsByName.Set(name, group)
	return group.id
}

func AddMember(gid GroupID, address string, weight int, metadata string) MemberID {
	runtime.AssertOriginCall()
	caller := runtime.OriginCaller()
	usernameOf(caller)
	group := getGroup(gid)
	if !group.HasPermission(caller, EditPermission) {
		panic("unauthorized to edit group")
	}
	user := users.ResolveAddress(chain.Address(address))
	if user == nil {
		panic("unknown address " + address)
	}
	mid := group.lastMemberID
	member := group.newMember(mid, chain.Address(address), weight, metadata)
	midkey := memberIDKey(mid)
	group.members.Set(midkey, member)
	mid++
	group.lastMemberID = mid
	return member.id
}

func DeleteGroup(gid GroupID) {
	runtime.AssertOriginCall()
	caller := runtime.OriginCaller()
	group := getGroup(gid)
	if !group.HasPermission(caller, DeletePermission) {
		panic("unauthorized to delete group")
	}
	group.deleteGroup()
}

func DeleteMember(gid GroupID, mid MemberID) {
	runtime.AssertOriginCall()
	caller := runtime.OriginCaller()
	group := getGroup(gid)
	if !group.HasPermission(caller, DeletePermission) {
		panic("unauthorized to delete member")
	}
	group.deleteMember(mid)
}
