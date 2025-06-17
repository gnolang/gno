package boards2

import (
	"errors"
	"std"
)

func validateBoardCreate(_ Permissions, args Args) error {
	name, ok := args[0].(string)
	if !ok {
		return errors.New("expected board name to be a string")
	}

	if err := checkBoardNameIsNotAddress(name); err != nil {
		return err
	}

	if err := checkBoardNameBelongsToCaller(name); err != nil {
		return err
	}
	return nil
}

func validateBoardRename(_ Permissions, args Args) error {
	newName, ok := args[2].(string)
	if !ok {
		return errors.New("expected new board name to be a string")
	}

	if err := checkBoardNameIsNotAddress(newName); err != nil {
		return err
	}

	if err := checkBoardNameBelongsToCaller(newName); err != nil {
		return err
	}
	return nil
}

func validateMemberInvite(perms Permissions, args Args) error {
	// Make sure that only owners invite other owners
	role, ok := args[1].(Role)
	if !ok {
		return errors.New("expected a valid new member role")
	}

	if role == RoleOwner {
		if !perms.HasRole(std.PreviousRealm().Address(), RoleOwner) {
			return errors.New("only owners are allowed to invite other owners")
		}
	}
	return nil
}

func validateRoleChange(perms Permissions, args Args) error {
	// Owners and Admins can change roles.
	// Admins should not be able to assign or remove the Owner role from members.
	if perms.HasRole(std.PreviousRealm().Address(), RoleAdmin) {
		role, ok := args[2].(Role)
		if !ok {
			return errors.New("expected a valid member role")
		}

		if role == RoleOwner {
			return errors.New("admins are not allowed to promote members to Owner")
		} else {
			member, ok := args[1].(std.Address)
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
