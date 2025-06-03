package group

import (
	"std"

	"gno.land/r/chaegnal/organizer"
)

const (
	memberRank = iota
	adminRank
)

// AcceptAwaitingMember is a wrapper to accept new members into friendgroup and DAO
func AcceptAwaitingMember(groupId, logged string, adr std.Address) {
	crossing()

	if !isAdmin(groupId, std.Address(logged)) {
		panic("only admin can do this action")
	}

	group := organizer.GetGroupFromId(groupId)
	if !group.WaitList.Has(adr.String()) {
		panic("user is not in the waitlist")
	}

	group.WaitList.Remove(adr.String())
	group.DAO.Members().Add(adr)
	group.Rank.Set(adr.String(), memberRank)
}

// RefuseAwaitingMember is a wrapper to refuse new members into friendgroup and DAO
func RefuseAwaitingMember(groupId, logged string, adr std.Address) {
	crossing()

	if !isAdmin(groupId, std.Address(logged)) {
		panic("only admin can do this action")
	}

	group := organizer.GetGroupFromId(groupId)
	group.WaitList.Remove(adr.String())
}

// RemoveMember is a wrapper to remove members from group and DAO
func RemoveMember(groupId, logged string, adr std.Address) {
	crossing()

	if !isAdmin(groupId, std.Address(logged)) {
		panic("only admin can do this action")
	}

	group := organizer.GetGroupFromId(groupId)
	if adr == group.Owner.Owner() {
		panic("owner cannot be removed")
	}

	group.DAO.Members().Remove(adr)
	group.Rank.Remove(adr.String())
}

// MakeAdmin is a wrapper to promote member to admin status
func MakeAdmin(groupId, logged string, adr std.Address) {
	crossing()

	if !isAdmin(groupId, std.Address(logged)) {
		panic("only admin can do this action")
	}

	group := organizer.GetGroupFromId(groupId)
	ok := group.Rank.Has(adr.String())
	if !ok {
		panic("user does not exist")
	}

	group.Rank.Set(adr.String(), adminRank)
}

// RemoveAdminRank is a wrapper to demote admin to member status
func RemoveAdminRank(groupId, logged string, adr std.Address) {
	crossing()

	if !isAdmin(groupId, std.Address(logged)) {
		panic("only admin can do this action")
	}

	group := organizer.GetGroupFromId(groupId)
	if adr == group.Owner.Owner() {
		panic("you cannot undo admin rights of the owner")
	}

	rank, ok := group.Rank.Get(adr.String())
	if !ok {
		panic("user does not exist")
	}

	if rank.(int) != adminRank {
		panic("user does not have admin right")
	}

	group.Rank.Set(adr.String(), memberRank)
}

func isAdmin(groupId string, adr std.Address) bool {
	group := organizer.GetGroupFromId(groupId)

	rank, ok := group.Rank.Get(adr.String())
	if !ok {
		panic("user does not exist")
	}

	return rank.(int) == adminRank
}
