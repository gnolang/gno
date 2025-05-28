package group

import (
	"std"

	"gno.land/r/chaegnal/organizer"
)

// AcceptAwaitingMember is a wrapper to accept new members into friendgroup and DAO
func AcceptAwaitingMember(groupId, logged string, adr std.Address) {
	crossing()
	group := organizer.GetGroupFromId(groupId)

	group.DAO.WhitelistUser(std.Address(logged), adr)
}

// RefuseAwaitingMember is a wrapper to refuse new members into friendgroup and DAO
func RefuseAwaitingMember(groupId, logged string, adr std.Address) {
	crossing()
	group := organizer.GetGroupFromId(groupId)

	group.DAO.RefuseAwaitingMember(std.Address(logged), adr)
}

// RemoveMember is a wrapper to remove members from group and DAO
func RemoveMember(groupId, logged string, adr std.Address) {
	crossing()
	group := organizer.GetGroupFromId(groupId)

	group.DAO.RemoveMember(std.Address(logged), adr)
}

// MakeAdmin is a wrapper to promote member to admin status
func MakeAdmin(groupId, logged string, adr std.Address) {
	crossing()
	group := organizer.GetGroupFromId(groupId)

	group.DAO.SetAdmin(std.Address(logged), adr)
}

// RemoveAdminRank is a wrapper to demote admin to member status
func RemoveAdminRank(groupId, logged string, adr std.Address) {
	crossing()
	group := organizer.GetGroupFromId(groupId)

	if adr == group.Owner.Owner() {
		panic("you cannot undo admin rights of the owner")
	}
	group.DAO.RemoveAdminRank(std.Address(logged), adr)
}
