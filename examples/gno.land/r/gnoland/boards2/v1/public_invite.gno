package boards2

import (
	"chain"
	"chain/runtime"
	"strings"
	"time"

	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/avl/v0"
)

// Invite contains a user invitation.
type Invite struct {
	// User is the user to invite.
	User address

	// Role is the optional role to assign to the user.
	Role boards.Role
}

// InviteMember adds a member to the realm or to a board.
//
// A role can optionally be specified to be assigned to the new member.
func InviteMember(_ realm, boardID boards.ID, user address, role boards.Role) {
	inviteMembers(boardID, Invite{
		User: user,
		Role: role,
	})
}

// InviteMembers adds one or more members to the realm or to a board.
//
// Board ID is only required when inviting a member to a specific board.
func InviteMembers(_ realm, boardID boards.ID, invites ...Invite) {
	inviteMembers(boardID, invites...)
}

// RequestInvite request to be invited to a board.
func RequestInvite(_ realm, boardID boards.ID) {
	assertMembersUpdateIsEnabled(boardID)

	if !runtime.PreviousRealm().IsUser() {
		panic("caller must be user")
	}

	// TODO: Request a fee (returned on accept) or registered user to avoid spam?
	// TODO: Make open invite requests optional (per board)

	board := mustGetBoard(boardID)
	user := runtime.PreviousRealm().Address()
	if board.Permissions.HasUser(user) {
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
func AcceptInvite(_ realm, boardID boards.ID, user address) {
	assertMembersUpdateIsEnabled(boardID)
	assertInviteRequestExists(boardID, user)

	board := mustGetBoard(boardID)
	if board.Permissions.HasUser(user) {
		panic("user is already a member")
	}

	caller := runtime.PreviousRealm().Address()
	invite := Invite{
		User: user,
		Role: RoleGuest,
	}
	args := boards.Args{caller, boardID, []Invite{invite}}
	board.Permissions.WithPermission(caller, PermissionMemberInvite, args, crossingFn(func() {
		assertMembersUpdateIsEnabled(boardID)

		invitee := user.String()
		requests, found := getInviteRequests(boardID)
		if !found || !requests.Has(invitee) {
			panic("invite request not found")
		}

		if board.Permissions.HasUser(user) {
			panic("user is already a member")
		}

		board.Permissions.SetUserRoles(user)
		requests.Remove(invitee)

		chain.Emit(
			"MembersInvited",
			"invitedBy", caller.String(),
			"boardID", board.ID.String(),
			"members", user.String()+":"+string(RoleGuest), // TODO: Support optional role assign
		)
	}))
}

// RevokeInvite revokes a board invite request.
func RevokeInvite(_ realm, boardID boards.ID, user address) {
	assertInviteRequestExists(boardID, user)

	board := mustGetBoard(boardID)
	caller := runtime.PreviousRealm().Address()
	args := boards.Args{boardID, user, RoleGuest}
	board.Permissions.WithPermission(caller, PermissionMemberInviteRevoke, args, crossingFn(func() {
		invitee := user.String()
		requests, found := getInviteRequests(boardID)
		if !found || !requests.Has(invitee) {
			panic("invite request not found")
		}

		requests.Remove(invitee)

		chain.Emit(
			"InviteRevoked",
			"revokedBy", caller.String(),
			"boardID", board.ID.String(),
			"user", user.String(),
		)
	}))
}

func inviteMembers(boardID boards.ID, invites ...Invite) {
	if len(invites) == 0 {
		panic("one or more user invites are required")
	}

	assertMembersUpdateIsEnabled(boardID)
	assertNoDuplicatedInvites(invites)

	perms := mustGetPermissions(boardID)
	caller := runtime.PreviousRealm().Address()
	args := boards.Args{caller, boardID, invites}
	perms.WithPermission(caller, PermissionMemberInvite, args, crossingFn(func() {
		assertMembersUpdateIsEnabled(boardID)

		users := make([]string, len(invites))
		for _, v := range invites {
			assertMemberAddressIsValid(v.User)

			if perms.HasUser(v.User) {
				panic("user is already a member: " + v.User.String())
			}

			// NOTE: Permissions implementation should check that role is valid
			perms.SetUserRoles(v.User, v.Role)
			users = append(users, v.User.String()+":"+string(v.Role))
		}

		chain.Emit(
			"MembersInvited",
			"invitedBy", caller.String(),
			"boardID", boardID.String(),
			"members", strings.Join(users, ","),
		)
	}))
}

func assertInviteRequestExists(boardID boards.ID, user address) {
	invitee := user.String()
	requests, found := getInviteRequests(boardID)
	if !found || !requests.Has(invitee) {
		panic("invite request not found")
	}
}

func assertNoDuplicatedInvites(invites []Invite) {
	if len(invites) == 1 {
		return
	}

	seen := make(map[address]struct{}, len(invites))
	for _, v := range invites {
		if _, found := seen[v.User]; found {
			panic("duplicated invite: " + v.User.String())
		}

		seen[v.User] = struct{}{}
	}
}
