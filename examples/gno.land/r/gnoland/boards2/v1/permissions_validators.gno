package boards2

import (
	"errors"

	"gno.land/p/gnoland/boards"

	"gno.land/r/sys/users"
)

// validateBoardCreate validates PermissionBoardCreate.
//
// Expected `args` values:
// 1. Caller address
// 2. Board name
// 3. Board ID
// 4. Is board listed
func validateBoardCreate(_ boards.Permissions, args boards.Args) error {
	caller, ok := args[0].(address)
	if !ok {
		return errors.New("expected a valid caller address")
	}

	name, ok := args[1].(string)
	if !ok {
		return errors.New("expected board name to be a string")
	}

	if err := checkBoardNameIsNotAddress(name); err != nil {
		return err
	}

	if err := checkBoardNameBelongsToAddress(caller, name); err != nil {
		return err
	}
	return nil
}

// validateBoardRename validates PermissionBoardRename.
//
// Expected `args` values:
// 1. Caller address
// 2. Board ID
// 3. Current board name
// 4. New board name
func validateBoardRename(_ boards.Permissions, args boards.Args) error {
	caller, ok := args[0].(address)
	if !ok {
		return errors.New("expected a valid caller address")
	}

	newName, ok := args[3].(string)
	if !ok {
		return errors.New("expected new board name to be a string")
	}

	if err := checkBoardNameIsNotAddress(newName); err != nil {
		return err
	}

	if err := checkBoardNameBelongsToAddress(caller, newName); err != nil {
		return err
	}
	return nil
}

// validateMemberInvite validates PermissionMemberInvite.
//
// Expected `args` values:
// 1. Caller address
// 2. Board ID
// 3. Invites
func validateMemberInvite(perms boards.Permissions, args boards.Args) error {
	caller, ok := args[0].(address)
	if !ok {
		return errors.New("expected a valid caller address")
	}

	invites, ok := args[2].([]Invite)
	if !ok {
		return errors.New("expected valid user invites")
	}

	// Make sure that only owners invite other owners
	callerIsOwner := perms.HasRole(caller, RoleOwner)
	for _, v := range invites {
		if v.Role == RoleOwner && !callerIsOwner {
			return errors.New("only owners are allowed to invite other owners")
		}
	}
	return nil
}

// validateRoleChange validates PermissionRoleChange.
//
// Expected `args` values:
// 1. Caller address
// 2. Board ID
// 3. Member address
// 4. Role
func validateRoleChange(perms boards.Permissions, args boards.Args) error {
	caller, ok := args[0].(address)
	if !ok {
		return errors.New("expected a valid caller address")
	}

	// Owners and Admins can change roles.
	// Admins should not be able to assign or remove the Owner role from members.
	if perms.HasRole(caller, RoleAdmin) {
		role, ok := args[3].(boards.Role)
		if !ok {
			return errors.New("expected a valid member role")
		}

		if role == RoleOwner {
			return errors.New("admins are not allowed to promote members to Owner")
		} else {
			member, ok := args[2].(address)
			if !ok {
				return errors.New("expected a valid member address")
			}

			if perms.HasRole(member, RoleOwner) {
				return errors.New("admins are not allowed to remove the Owner role")
			}
		}
	}
	return nil
}

func checkBoardNameIsNotAddress(s string) error {
	if address(s).IsValid() {
		return errors.New("addresses are not allowed as board name")
	}
	return nil
}

func checkBoardNameBelongsToAddress(owner address, name string) error {
	// When the board name is the name of a registered user
	// check that caller is the owner of the name.
	user, _ := users.ResolveName(name)
	if user != nil && user.Addr() != owner {
		return errors.New("board name is a user name registered to a different user")
	}
	return nil
}
