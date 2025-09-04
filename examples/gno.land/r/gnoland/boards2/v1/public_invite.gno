package boards2

import (
	"chain"
	"chain/runtime"
	"time"

	"gno.land/p/demo/avl"
)

// InviteMember adds a member to the realm or to a boards.
//
// A role can optionally be specified to be assigned to the new member.
// Board ID is only required when inviting a member to a board.
func InviteMember(_ realm, boardID BoardID, user address, role Role) {
	assertUserAddressIsValid(user)
	assertMembersUpdateIsEnabled(boardID)

	perms := mustGetPermissions(boardID)
	if perms.HasUser(user) {
		panic("user is already a member")
	}

	caller := runtime.PreviousRealm().Address()
	args := Args{boardID, user, role}
	perms.WithPermission(caller, PermissionMemberInvite, args, func(Args) {
		assertMembersUpdateIsEnabled(boardID)

		if perms.HasUser(user) {
			panic("user is already a member")
		}

		if err := perms.SetUserRoles(user, role); err != nil {
			panic(err)
		}

		chain.Emit(
			"MemberInvited",
			"invitedBy", caller.String(),
			"boardID", boardID.String(),
			"user", user.String(),
			"role", string(role),
		)
	})
}

// RequestInvite request to be invited to a board.
func RequestInvite(_ realm, boardID BoardID) {
	assertMembersUpdateIsEnabled(boardID)

	if !runtime.PreviousRealm().IsUser() {
		panic("caller must be user")
	}

	// TODO: Request a fee (returned on accept) or registered user to avoid spam?

	board := mustGetBoard(boardID)
	user := runtime.PreviousRealm().Address()
	if board.perms.HasUser(user) {
		panic("caller is already a member")
	}

	invitee := user.String()
	requests, found := getInviteRequests(boardID)
	if !found {
		requests = avl.NewTree()
		requests.Set(invitee, time.Now())
		gInviteRequests.Set(boardID.Key(), requests)
		return
	}

	if requests.Has(invitee) {
		panic("invite request already exists")
	}

	requests.Set(invitee, time.Now())
}

// AcceptInvite accepts a board invite request.
func AcceptInvite(_ realm, boardID BoardID, user address) {
	assertMembersUpdateIsEnabled(boardID)
	assertInviteRequestExists(boardID, user)

	board := mustGetBoard(boardID)
	if board.perms.HasUser(user) {
		panic("user is already a member")
	}

	caller := runtime.PreviousRealm().Address()
	args := Args{boardID, user, RoleGuest}
	board.perms.WithPermission(caller, PermissionMemberInvite, args, func(Args) {
		assertMembersUpdateIsEnabled(boardID)

		invitee := user.String()
		requests, found := getInviteRequests(boardID)
		if !found || !requests.Has(invitee) {
			panic("invite request not found")
		}

		board := mustGetBoard(boardID)
		if board.perms.HasUser(user) {
			panic("user is already a member")
		}

		if err := board.perms.SetUserRoles(user); err != nil {
			panic(err)
		}

		requests.Remove(invitee)

		chain.Emit(
			"MemberInvited",
			"invitedBy", caller.String(),
			"boardID", boardID.String(),
			"user", user.String(),
		)
	})
}

// RevokeInvite revokes a board invite request.
func RevokeInvite(_ realm, boardID BoardID, user address) {
	assertInviteRequestExists(boardID, user)

	board := mustGetBoard(boardID)
	caller := runtime.PreviousRealm().Address()
	args := Args{boardID, user, RoleGuest}
	board.perms.WithPermission(caller, PermissionMemberInviteRevoke, args, func(Args) {
		invitee := user.String()
		requests, found := getInviteRequests(boardID)
		if !found || !requests.Has(invitee) {
			panic("invite request not found")
		}

		requests.Remove(invitee)

		chain.Emit(
			"InviteRevoked",
			"revokedBy", caller.String(),
			"boardID", boardID.String(),
			"user", user.String(),
		)
	})
}

func assertInviteRequestExists(boardID BoardID, user address) {
	invitee := user.String()
	requests, found := getInviteRequests(boardID)
	if !found || !requests.Has(invitee) {
		panic("invite request not found")
	}
}
