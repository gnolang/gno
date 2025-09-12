package boards2

import (
	"errors"
	"std"
)

// validateBoardCreate validates PermissionBoardCreate.
//
// Expected `args` values:
// 1. Caller address
// 2. Board name
// 3. Board ID
// 4. Is board listed
func validateBoardCreate(_ Permissions, args Args) error {
	caller, ok := args[0].(std.Address)
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
func validateBoardRename(_ Permissions, args Args) error {
	caller, ok := args[0].(std.Address)
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
// 3. User address
// 4. Role
func validateMemberInvite(perms Permissions, args Args) error {
	caller, ok := args[0].(std.Address)
	if !ok {
		return errors.New("expected a valid caller address")
	}

	// Make sure that only owners invite other owners
	role, ok := args[3].(Role)
	if !ok {
		return errors.New("expected a valid new member role")
	}

	if role == RoleOwner {
		if !perms.HasRole(caller, RoleOwner) {
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
func validateRoleChange(perms Permissions, args Args) error {
	caller, ok := args[0].(std.Address)
	if !ok {
		return errors.New("expected a valid caller address")
	}

	// Owners and Admins can change roles.
	// Admins should not be able to assign or remove the Owner role from members.
	if perms.HasRole(caller, RoleAdmin) {
		role, ok := args[3].(Role)
		if !ok {
			return errors.New("expected a valid member role")
		}

		if role == RoleOwner {
			return errors.New("admins are not allowed to promote members to Owner")
		} else {
			member, ok := args[2].(std.Address)
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
