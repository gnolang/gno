package permissions

import (
	"chain/runtime/unsafe"

	"gno.land/p/gnoland/boards/exts/permissions"

	"gno.land/r/gnoland/boards2/v1"
)

var (
	// Admin is the address of the current realm admin.
	Admin address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh" // GovDAO T1 multisig

	// OpenAccountAmount is the ugnot required in an account for open board interactions.
	// This requirement is only for non board members.
	OpenAccountAmount int64 = 3_000_000_000
)

// SetAdmin sets a new realm admin address.
func SetAdmin(_ realm, admin address) {
	caller := unsafe.PreviousRealm().Address()
	if caller != Admin {
		panic("unauthorized")
	}

	Admin = admin
}

// SetOpenAccountAmount sets the required amount in ugnot for open board interactions.
func SetOpenAccountAmount(_ realm, amount int64) {
	caller := unsafe.PreviousRealm().Address()
	if caller != Admin {
		panic("unauthorized")
	}

	if amount < 0 {
		panic("invalid amount")
	}

	OpenAccountAmount = amount
}

// New creates custom permissions for open boards.
func New(owner address) *permissions.Permissions {
	perms := permissions.New(
		permissions.UseSingleUserRole(),
		permissions.WithSuperRole(boards2.RoleOwner),
	)
	perms.SetPublicPermissions(
		boards2.PermissionThreadCreate,
		boards2.PermissionThreadRepost,
		boards2.PermissionReplyCreate,
	)
	perms.AddRole(
		boards2.RoleAdmin,
		boards2.PermissionBoardRename,
		boards2.PermissionBoardFlaggingUpdate,
		boards2.PermissionMemberInvite,
		boards2.PermissionMemberInviteRevoke,
		boards2.PermissionMemberRemove,
		boards2.PermissionThreadCreate,
		boards2.PermissionThreadEdit,
		boards2.PermissionThreadDelete,
		boards2.PermissionThreadRepost,
		boards2.PermissionThreadFlag,
		boards2.PermissionThreadFreeze,
		boards2.PermissionReplyCreate,
		boards2.PermissionReplyDelete,
		boards2.PermissionReplyFlag,
		boards2.PermissionRoleChange,
		boards2.PermissionUserBan,
		boards2.PermissionUserUnban,
	)
	perms.AddRole(
		boards2.RoleModerator,
		boards2.PermissionThreadCreate,
		boards2.PermissionThreadEdit,
		boards2.PermissionThreadRepost,
		boards2.PermissionThreadFlag,
		boards2.PermissionReplyCreate,
		boards2.PermissionReplyFlag,
		boards2.PermissionUserBan,
		boards2.PermissionUserUnban,
	)
	perms.AddRole(
		boards2.RoleGuest,
		boards2.PermissionThreadRepost,
		boards2.PermissionReplyCreate,
	)
	perms.SetUserRoles(owner, boards2.RoleOwner)
	perms.ValidateFunc(boards2.PermissionBoardRename, validateOpenBoardRename)
	perms.ValidateFunc(boards2.PermissionMemberInvite, validateOpenMemberInvite)
	perms.ValidateFunc(boards2.PermissionRoleChange, validateOpenRoleChange)
	perms.ValidateFunc(boards2.PermissionThreadCreate, validateOpenThreadCreate)
	perms.ValidateFunc(boards2.PermissionReplyCreate, validateOpenReplyCreate)
	return perms
}


