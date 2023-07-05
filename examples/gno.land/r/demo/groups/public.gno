package groups

import (
	"std"

	"gno.land/r/demo/users"
)

//----------------------------------------
// Public facing functions

func GetGroupNameFromID(gid GroupID) (string, bool) {
	groupI, exists := gGroups.Get(groupIDKey(gid))
	if !exists {
		return "", false
	}
	return groupI.(*Group).name, true
}

func GetGroupIDFromName(name string) (GroupID, bool) {
	groupI, exists := gGroupsByName.Get(name)
	if !exists {
		return 0, false
	}
	return groupI.(*Group).id, true
}

func CreateGroup(name string) GroupID {
	// std.AssertOriginCall()
	caller := std.PrevRealm().Addr()
	// usernameOf(caller)
	url := "/r/demo/groups:" + name
	group := newGroup(url, name, caller)
	gidkey := groupIDKey(group.id)
	gGroups.Set(gidkey, group)
	gGroupsByName.Set(name, group)
	return group.id
}

func AddMember(gid GroupID, address string, weight int, metadata string) MemberID {
	// std.AssertOriginCall()
	caller := std.PrevRealm().Addr()
	// usernameOf(caller)
	group := getGroup(gid)
	if !group.HasPermission(caller, EditPermission) {
		panic("unauthorized to edit group")
	}
	/*
		user := users.GetUserByAddress(std.Address(address))
		if user == nil {
			panic("unknown address " + address)
		}
	*/
	mid := group.lastMemberID
	member := group.newMember(mid, std.Address(address), weight, metadata)
	midkey := memberIDKey(mid)
	group.members.Set(midkey, member)
	group.membersByAddress.Set(address, member)
	group.totalWeight += weight
	mid++
	group.lastMemberID = mid
	return member.id
}

func DeleteGroup(gid GroupID) {
	// std.AssertOriginCall()
	caller := std.PrevRealm().Addr()
	group := getGroup(gid)
	if !group.HasPermission(caller, DeletePermission) {
		panic("unauthorized to delete group")
	}
	group.deleteGroup()
}

func DeleteMember(gid GroupID, mid MemberID) {
	// std.AssertOriginCall()
	caller := std.PrevRealm().Addr()
	group := getGroup(gid)
	if !group.HasPermission(caller, DeletePermission) {
		panic("unauthorized to delete member")
	}
	group.deleteMember(mid)
}

func GetMemberWeightByAddress(gid GroupID, addr std.Address) int {
	group := getGroup(gid)
	if group == nil {
		return 0
	}
	member, ok := group.membersByAddress.Get(string(addr))
	if !ok {
		return 0
	}
	return member.(*Member).weight
}

func GetGroupTotalWeight(gid GroupID) int {
	group := getGroup(gid)
	if group == nil {
		return 0
	}
	return group.totalWeight
}

func GetMembers(gid GroupID) []Member {
	group := getGroup(gid)
	if group == nil {
		return nil
	}
	members := []Member{}
	group.members.Iterate("", "", func(key string, value interface{}) bool {
		members = append(members, *(value.(*Member)))
		return false
	})
	return members
}
