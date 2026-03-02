package group

import (
	"std"

	"gno.land/p/demo/ownable"

	"gno.land/r/chaegnal/organizer"
)

// DeleteFriendGroup allows the owner to delete the friendgroup
// Panics if anyone but the owner try to delete the group
func DeleteFriendGroup(cur realm, groupId, logged string) {
	group := organizer.GetGroupFromId(groupId)

	if group.Owner.Owner() != std.Address(logged) {
		panic("only the owner can delete the group")
	}

	organizer.ListGroups.Groups.Remove(group.ID.String())
}

// SwitchOwner enables giving out ownership of the group
// Panics if the caller is not the owner or if the new owner is not an admin
func SwitchOwner(cur realm, groupId, logged string, adr std.Address) {
	group := organizer.GetGroupFromId(groupId)

	if group.Owner.Owner() != std.Address(logged) {
		panic("only the owner give ownership away")
	}

	ok := isAdmin(groupId, adr)
	if !ok {
		panic("the new owner must be an admin")
	}

	group.Owner = ownable.NewWithAddress(adr)
	organizer.ListGroups.Groups.Set(group.ID.String(), group)
}
