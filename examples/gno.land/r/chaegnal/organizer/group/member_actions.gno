package group

import (
	"std"

	"gno.land/r/chaegnal/organizer"
)

// acceptAwaitingMember accepts new members into friendgroup and DAO
func acceptAwaitingMember(cur realm, groupId, logged string, adr std.Address) error {
	if !isAdmin(groupId, std.Address(logged)) {
		panic("only admin can do this action")
	}

	group := organizer.GetGroupFromId(groupId)
	if !group.WaitList.Has(adr.String()) {
		panic("user is not in the waitlist")
	}

	group.WaitList.Remove(adr.String())
	group.DAO.Members().Add(adr)
	group.Rank.Set(adr.String(), organizer.MemberRank)
	return nil
}

// refuseAwaitingMember refuses new members into friendgroup and DAO
func refuseAwaitingMember(cur realm, groupId, logged string, adr std.Address) error {
	if !isAdmin(groupId, std.Address(logged)) {
		panic("only admin can do this action")
	}

	group := organizer.GetGroupFromId(groupId)
	group.WaitList.Remove(adr.String())
	return nil
}

// removeMember removes members from group and DAO
func removeMember(cur realm, groupId, logged string, adr std.Address) error {
	if !isAdmin(groupId, std.Address(logged)) {
		panic("only admin can do this action")
	}

	group := organizer.GetGroupFromId(groupId)
	if adr == group.Owner.Owner() {
		panic("owner cannot be removed")
	}

	group.DAO.Members().Remove(adr)
	group.Rank.Remove(adr.String())
	return nil
}

// makeAdmin promotes member to admin status
func makeAdmin(cur realm, groupId, logged string, adr std.Address) error {
	if !isAdmin(groupId, std.Address(logged)) {
		panic("only admin can do this action")
	}

	group := organizer.GetGroupFromId(groupId)
	ok := group.Rank.Has(adr.String())
	if !ok {
		panic("user does not exist")
	}

	group.Rank.Set(adr.String(), organizer.AdminRank)
	return nil
}

// rmoveAdminRank demotes admin to member status
func removeAdminRank(cur realm, groupId, logged string, adr std.Address) error {
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

	if rank.(int) != organizer.AdminRank {
		panic("user does not have admin right")
	}

	group.Rank.Set(adr.String(), organizer.MemberRank)
	return nil
}

func isAdmin(groupId string, adr std.Address) bool {
	group := organizer.GetGroupFromId(groupId)

	rank, ok := group.Rank.Get(adr.String())
	if !ok {
		panic("user does not exist")
	}

	return rank.(int) == organizer.AdminRank
}
