package permissions

import (
	"chain/banker"
	"errors"

	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/ufmt/v0"

	"gno.land/r/gnoland/boards2/v1"
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

	invites, ok := args[2].([]boards2.Invite)
	if !ok {
		return errors.New("expected valid user invites")
	}

	// Make sure that only owners invite other owners
	callerIsOwner := perms.HasRole(caller, boards2.RoleOwner)
	for _, v := range invites {
		if v.Role == boards2.RoleOwner && !callerIsOwner {
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
	if perms.HasRole(caller, boards2.RoleAdmin) {
		role, ok := args[3].(boards.Role)
		if !ok {
			return errors.New("expected a valid member role")
		}

		if role == boards2.RoleOwner {
			return errors.New("admins are not allowed to promote members to Owner")
		} else {
			member, ok := args[2].(address)
			if !ok {
				return errors.New("expected a valid member address")
			}

			if perms.HasRole(member, boards2.RoleOwner) {
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

	// All board members can create threads
	if perms.HasUser(caller) {
		return nil
	}

	// Require non members to have some GNOT in their accounts
	if err := checkAccountHasAmount(caller, OpenAccountAmount); err != nil {
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

	// Require non members to have some GNOT in their accounts
	if err := checkAccountHasAmount(caller, OpenAccountAmount); err != nil {
		return ufmt.Errorf("caller is not allowed to comment: %s", err)
	}
	return nil
}

func checkAccountHasAmount(addr address, amount int64) error {
	bnk := banker.NewReadonlyBanker()
	coins := bnk.GetCoins(addr)
	if coins.AmountOf("ugnot") < OpenAccountAmount {
		amount = amount / 1_000_000 // ugnot -> GNOT
		return ufmt.Errorf("account amount is lower than %d GNOT", amount)
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


