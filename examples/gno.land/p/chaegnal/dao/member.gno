package dao

import (
	"std"

	"gno.land/p/demo/avl/rotree"
)

type WhitelistedUser struct {
	Address std.Address
	Rank    string
}

const (
	adminRank  = "admin"
	memberRank = "member"
)

// Verify that a user is in the whitelist
func (dao *DAO) IsMember(adr std.Address) (WhitelistedUser, bool) {
	raw, ok := dao.Whitelist.Get(adr.String())
	if !ok {
		return WhitelistedUser{}, false
	}

	return raw.(WhitelistedUser), true
}

// Verify that a user is in the whitelist and possesses the admin rank
func (dao *DAO) IsAdmin(adr std.Address) bool {
	member, isMember := dao.IsMember(adr)

	return isMember && member.Rank == adminRank
}

// Verify if a user is in the waitlist
func (dao *DAO) IsPending(adr std.Address) bool {
	_, ok := dao.AwaitingApproval.Get(adr.String())

	return ok
}

// Getters

func (dao *DAO) GetWhitelisted() rotree.IReadOnlyTree {
	return rotree.Wrap(dao.Whitelist, nil)
}

func (dao *DAO) GetAwaiting() rotree.IReadOnlyTree {
	return rotree.Wrap(dao.AwaitingApproval, nil)
}

// Set user into the waitlist
// Verify if the user is already in the waitlist or in the whitelist
func (dao *DAO) PendingApproval(adr std.Address) {
	_, err := dao.AwaitingApproval.Get(adr.String())
	if err {
		panic("already awaiting approval")
	}

	_, err = dao.IsMember(adr)
	if err {
		panic("already whitelisted")
	}

	dao.AwaitingApproval.Set(adr.String(), true)
}

// Move user from the waitlist to the whitelist
// Verify that the action is done by an admin
// Verify that the user is in the waitlist and not already in the whitelist
func (dao *DAO) WhitelistUser(caller, adr std.Address) {
	if !dao.IsAdmin(caller) {
		panic("only admins can whitelist members")
	}

	_, ok := dao.AwaitingApproval.Get(adr.String())
	if !ok {
		panic("this address is not awaiting approval")
	}

	_, isMember := dao.IsMember(adr)
	if isMember {
		panic("already whitelisted")
	}

	dao.whitelistMemberNoApproval(adr)
	dao.AwaitingApproval.Remove(adr.String())
}

// Remove user from the waitlist
// Verify that the action is done by an admin
func (dao *DAO) RefuseAwaitingMember(caller, adr std.Address) {
	if !dao.IsAdmin(caller) {
		panic("only admins can refuse members")
	}

	dao.AwaitingApproval.Remove(adr.String())
}

// Allow admin to set a user as an admin
func (dao *DAO) SetAdmin(caller, adr std.Address) {
	if !dao.IsAdmin(caller) {
		panic("only admins can promote members")
	}

	dao.defaultAdmin(adr)
}

// Demote admin to member
// Verify is the action is made by an admin, if the member exist and is a admin
// Bans auto-demotion
func (dao *DAO) RemoveAdminRank(caller, adr std.Address) {
	if !dao.IsAdmin(caller) {
		panic("only admins can demote members")
	}

	member, _ := dao.IsMember(adr)
	if !dao.IsAdmin(adr) {
		panic("member is not an admin")
	}

	if caller == adr {
		panic("cannot demote yourself")
	}

	member.Rank = memberRank
	dao.Whitelist.Set(adr.String(), member)
}

// Remove member from whitelist
// Can only be done by an admin on an existing member
// Bans auto-removal
func (dao *DAO) RemoveMember(caller, adr std.Address) {
	if !dao.IsAdmin(caller) {
		panic("must be an admin to remove a member")
	}

	member, ok := dao.IsMember(adr)
	if !ok {
		panic("not a member")
	}

	if member.Address == caller {
		panic("you cannot remove yourself")
	}

	dao.Whitelist.Remove(adr.String())
}

// Let member leave the group
// Bans admin from leaving if no other admin exists
func (dao *DAO) LeaveDao(caller std.Address) {
	var adminCounter int

	_, ok := dao.IsMember(caller)
	if !ok {
		panic("not a member")
	}

	if dao.IsAdmin(caller) {
		adminCounter = dao.countAdmins()
		if adminCounter <= 1 {
			panic("you are the last admin: assign another admin before leaving")
		}
	}

	dao.Whitelist.Remove(caller.String())
}

// Helper functions

// Set a default admin at the creation of the DAO
// Private function to maximize security
func (dao *DAO) defaultAdmin(adr std.Address) {
	member, ok := dao.IsMember(adr)
	if dao.IsAdmin(adr) {
		panic("member is already an admin")
	}

	if !ok {
		panic("not a member")
	}

	member.Rank = adminRank
	dao.Whitelist.Set(adr.String(), member)
}

// Set a user into the whitelist with a default member rank
func (dao *DAO) whitelistMemberNoApproval(adr std.Address) {
	newWhitelisted := WhitelistedUser{
		Address: adr,
		Rank:    memberRank,
	}

	dao.Whitelist.Set(adr.String(), newWhitelisted)
}

func (dao *DAO) countAdmins() int {
	var adminCounter int

	dao.Whitelist.Iterate("", "", func(_ string, val any) bool {
		user := val.(WhitelistedUser)

		if user.Rank == adminRank {
			adminCounter++
		}

		return false
	})

	return adminCounter
}
