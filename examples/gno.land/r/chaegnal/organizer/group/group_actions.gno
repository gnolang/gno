package group

import (
	"std"
	"strconv"

	"gno.land/p/demo/ownable"

	"gno.land/r/chaegnal/organizer"
)

// Allows the owner to delete the friendgroup
func DeleteFriendGroup(groupId, logged string) {
	crossing()
	group := organizer.GetGroupFromId(groupId)

	if group.Owner.Owner() != std.Address(logged) {
		panic("only the owner can delete the group")
	}

	organizer.ListGroups.Groups.Remove(strconv.Itoa(group.ID))
}

// Enables giving out ownership of the group
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
