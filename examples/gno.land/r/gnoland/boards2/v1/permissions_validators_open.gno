package boards2

import (
	"chain/banker"
	"errors"

	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/ufmt"

	"gno.land/r/sys/users"
)

// validateOpenBoardRename validates PermissionBoardRename.
//
// Expected `args` values:
// 1. Caller address
// 2. Board ID
// 3. Current board name
// 4. New board name
func validateOpenBoardRename(_ boards.Permissions, args boards.Args) error {
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

// validateOpenMemberInvite validates PermissionMemberInvite.
//
// Expected `args` values:
// 1. Caller address
// 2. Board ID
// 3. Invites
func validateOpenMemberInvite(perms boards.Permissions, args boards.Args) error {
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

// validateOpenRoleChange validates PermissionRoleChange.
//
// Expected `args` values:
// 1. Caller address
// 2. Board ID
// 3. Member address
// 4. Role
func validateOpenRoleChange(perms boards.Permissions, args boards.Args) error {
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

// validateOpenThreadCreate validates PermissionThreadCreate.
//
// Expected `args` values:
// 1. Caller address
// 2. Board ID
// 3. Thread ID
// 4. Title
// 5. Body
func validateOpenThreadCreate(perms boards.Permissions, args boards.Args) error {
	caller, ok := args[0].(address)
	if !ok {
		return errors.New("expected a valid caller address")
	}

	// Owners and admins can create threads without a registered username
	if perms.HasRole(caller, RoleOwner) || perms.HasRole(caller, RoleAdmin) {
		return nil
	}

	// Require registered usernames to non members
	if users.ResolveAddress(caller) == nil {
		return errors.New("a registered username is required to create threads on open boards")
	}

	// Require non members to have some GNOT in their accounts
	if err := checkAccountHasAmount(caller, gOpenAccountAmount); err != nil {
		return ufmt.Errorf("caller is not allowed to create threads: %s", err)
	}
	return nil
}

// validateOpenReplyCreate validates PermissionReplyCreate.
//
// Expected `args` values:
// 1. Caller address
// 2. Board ID
// 3. Thread ID
// 4. Parent ID
// 5. Reply ID
// 6. Body
func validateOpenReplyCreate(perms boards.Permissions, args boards.Args) error {
	caller, ok := args[0].(address)
	if !ok {
		return errors.New("expected a valid caller address")
	}

	// All board members can reply
	if perms.HasUser(caller) {
		return nil
	}

	// Require registered usernames to non members
	if users.ResolveAddress(caller) == nil {
		return errors.New("a registered username is required to comment on open boards")
	}

	// Require non members to have some GNOT in their accounts
	if err := checkAccountHasAmount(caller, gOpenAccountAmount); err != nil {
		return ufmt.Errorf("caller is not allowed to comment: %s", err)
	}
	return nil
}

func checkAccountHasAmount(addr address, amount int64) error {
	bnk := banker.NewBanker(banker.BankerTypeReadonly)
	coins := bnk.GetCoins(addr)
	if coins.AmountOf("ugnot") < gOpenAccountAmount {
		amount = amount / 1_000_000 // ugnot -> GNOT
		return ufmt.Errorf("account amount is lower than %d GNOT", amount)
	}
	return nil
}
