package boards2

import (
	"chain"
	"chain/runtime"
	"strings"
	"time"

	"gno.land/p/nt/avl"
)

// Invite contains a user invitation.
type Invite struct {
	// User is the user to invite.
	User address

	// Role is the optional role to assign to the user.
	Role Role
}

// InviteMember adds a member to the realm or to a board.
//
// A role can optionally be specified to be assigned to the new member.
func InviteMember(_ realm, boardID BoardID, user address, role Role) {
	inviteMembers(boardID, Invite{
		User: user,
		Role: role,
	})
}

// InviteMembers adds one or more members to the realm or to a board.
//
// Board ID is only required when inviting a member to a specific board.
func InviteMembers(_ realm, boardID BoardID, invites ...Invite) {
	inviteMembers(boardID, invites...)
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
	invite := Invite{
		User: user,
		Role: RoleGuest,
	}
	args := Args{caller, boardID, []Invite{invite}}
	board.perms.WithPermission(cross, caller, PermissionMemberInvite, args, func(realm) {
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

		board.perms.SetUserRoles(cross, user)
		requests.Remove(invitee)

		chain.Emit(
			"MembersInvited",
			"invitedBy", caller.String(),
			"boardID", boardID.String(),
			"members", user.String()+":"+string(RoleGuest), // TODO: Support optional role assign
		)
	})
}

// RevokeInvite revokes a board invite request.
func RevokeInvite(_ realm, boardID BoardID, user address) {
	assertInviteRequestExists(boardID, user)

	board := mustGetBoard(boardID)
	caller := runtime.PreviousRealm().Address()
	args := Args{boardID, user, RoleGuest}
	board.perms.WithPermission(cross, caller, PermissionMemberInviteRevoke, args, func(realm) {
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

func inviteMembers(boardID BoardID, invites ...Invite) {
	assertMembersUpdateIsEnabled(boardID)

	if len(invites) == 0 {
		panic("one or more user invites are required")
	}

	perms := mustGetPermissions(boardID)
	caller := runtime.PreviousRealm().Address()
	args := Args{caller, boardID, invites}
	perms.WithPermission(cross, caller, PermissionMemberInvite, args, func(realm) {
		assertMembersUpdateIsEnabled(boardID)

		perms := mustGetPermissions(boardID)
		users := make([]string, len(invites))
		for _, v := range invites {
			assertMemberAddressIsValid(v.User)

			if perms.HasUser(v.User) {
				panic("user is already a member: " + v.User.String())
			}

			// NOTE: Permissions implementation should check that role is valid
			perms.SetUserRoles(cross, v.User, v.Role)
			users = append(users, v.User.String()+":"+string(v.Role))
		}

		chain.Emit(
			"MembersInvited",
			"invitedBy", caller.String(),
			"boardID", boardID.String(),
			"members", strings.Join(users, ","),
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
