package group

import (
	"std"
	"strconv"

	"gno.land/p/demo/ownable"

	"gno.land/r/chaegnal/organizer"
)

// DeleteFriendGroup allows the owner to delete the friendgroup
// Panics if anyone but the owner try to delete the group
func DeleteFriendGroup(groupId, logged string) {
	crossing()
	group := organizer.GetGroupFromId(groupId)

	if group.Owner.Owner() != std.Address(logged) {
		panic("only the owner can delete the group")
	}

	organizer.ListGroups.Groups.Remove(strconv.Itoa(group.ID))
}

// SwitchOwner enables giving out ownership of the group
// Panics if the caller is not the owner or if the new owner is not an admin
func SwitchOwner(groupId, logged string, adr std.Address) {
	crossing()
	group := organizer.GetGroupFromId(groupId)

	if group.Owner.Owner() != std.Address(logged) {
		panic("only the owner give ownership away")
	}

	ok := group.DAO.IsAdmin(adr)
	if !ok {
		panic("the new owner must be an admin")
	}

	group.Owner = ownable.NewWithAddress(adr)
	organizer.ListGroups.Groups.Set(strconv.Itoa(group.ID), group)
}
