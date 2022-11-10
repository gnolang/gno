package groups

import (
	"std"
	"strconv"

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

func CreateGroup(name string) GroupID {

	//m := membInpParser(memberList)
	std.AssertOriginCall()
	gid := incGetGroupID()
	caller := std.GetOrigCaller()
	if usernameOf(caller) == "" {
		panic("unauthorized")
	}
	url := "/r/demo/groups:" + name
	group := newGroup(gid, url, name, caller)
	gidkey := groupIDKey(gid)
	gGroups.Set(gidkey, group)
	gGroupsByName.Set(name, group)

	/*for i, j := range m {
		mid := group.lastMemberID
		//check if user address is valid and exist
		user := users.GetUserByAddress(std.Address(j.address))
		if user == nil {
			continue
		}
		member := newMember(mid, gid, j.address, j.weight, j.metadata)
		midkey := memberIDKey(mid)
		group.members.Set(midkey, member)
		mid++
		group.lastMemberID = mid
	}*/

	return group.id
}

func AddMember(gid GroupID, address string, weight int, metadata string) MemberID {

	std.AssertOriginCall()
	caller := std.GetOrigCaller()
	if usernameOf(caller) == "" {
		panic("unauthorized")
	}

	group := getGroup(gid)
	if group == nil {
		panic("group not exist")
	}

	if !group.HasPermission(caller, EditPermission) {
		panic("unauthorized to edit group")
	}

	user := users.GetUserByAddress(std.Address(address))
	if user == nil {
		panic("unknown address " + address)
	}

	mid := group.lastMemberID
	member := newMember(mid, gid, address, weight, metadata)
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
	if group == nil {
		panic("group not exist")
	}
	if !group.HasPermission(caller, DeletePermission) {
		panic("unauthorized to delete group")
	}
	group.deleteGroup(gid)
}

func DeleteMember(gid GroupID, mid MemberID) {
	std.AssertOriginCall()
	caller := std.GetOrigCaller()
	group := getGroup(gid)
	if group == nil {
		panic("group not exist")
	}
	if !group.HasPermission(caller, DeletePermission) {
		panic("unauthorized to delete member")
	}
	group.deleteMember(gid, mid)
}
