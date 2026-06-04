package boards2

import (
	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/permissions"
)

// List of Boards2 member roles.
const (
	RoleOwner     boards.Role = "owner"
	RoleAdmin                 = "admin"
	RoleModerator             = "moderator"
	RoleGuest                 = "guest"
)

// PermissionCustom defines an initial value for custom board permissions.
// When a board defines custom permissions it must starts from a value
// greater or equal than PermissionCustom.
//
// Custom permissions definition example:
//
//	const (
//	  PermissionCustom1 boards.Permission = iota + boards2.PermissionCustom
//	  PermissionCustom2
//	  PermissionCustom3
//	)
const PermissionCustom boards.Permission = 200

// List of Boards2 permissions.
const (
	PermissionBoardCreate boards.Permission = iota
	PermissionBoardFlaggingUpdate
	PermissionBoardFreeze
	PermissionBoardRename
	PermissionMemberInvite
	PermissionMemberInviteRevoke
	PermissionMemberRemove
	PermissionPermissionsUpdate
	PermissionRealmHelpChange
	PermissionRealmLock
	PermissionRealmNotice
	PermissionAccountRequiredAmountChange
	PermissionReplyCreate
	PermissionReplyDelete
	PermissionReplyFlag
	PermissionRoleChange
	PermissionThreadCreate
	PermissionThreadDelete
	PermissionThreadEdit
	PermissionThreadFlag
	PermissionThreadFreeze
	PermissionThreadRepost
	PermissionUserBan
	PermissionUserUnban
)

func createBasicBoardPermissions(owner address) *permissions.Permissions {
	perms := permissions.New(
		permissions.UseSingleUserRole(),
		permissions.WithSuperRole(RoleOwner),
	)
	perms.AddRole(
		RoleAdmin,
		PermissionBoardRename,
		PermissionBoardFlaggingUpdate,
		PermissionMemberInvite,
		PermissionMemberInviteRevoke,
		PermissionMemberRemove,
		PermissionThreadCreate,
		PermissionThreadEdit,
		PermissionThreadDelete,
		PermissionThreadRepost,
		PermissionThreadFlag,
		PermissionThreadFreeze,
		PermissionReplyCreate,
		PermissionReplyDelete,
		PermissionReplyFlag,
		PermissionRoleChange,
		PermissionUserBan,
		PermissionUserUnban,
	)
	perms.AddRole(
		RoleModerator,
		PermissionThreadCreate,
		PermissionThreadEdit,
		PermissionThreadRepost,
		PermissionThreadFlag,
		PermissionReplyCreate,
		PermissionReplyFlag,
		PermissionUserBan,
		PermissionUserUnban,
	)
	perms.AddRole(
		RoleGuest,
		PermissionThreadCreate,
		PermissionThreadRepost,
		PermissionReplyCreate,
	)
	perms.SetUserRoles(owner, RoleOwner)
	perms.ValidateFunc(PermissionBoardRename, validateBasicBoardRename)
	perms.ValidateFunc(PermissionMemberInvite, validateBasicMemberInvite)
	perms.ValidateFunc(PermissionRoleChange, validateBasicRoleChange)
	return perms
}

func createOpenBoardPermissions(owner address) *permissions.Permissions {
	perms := permissions.New(
		permissions.UseSingleUserRole(),
		permissions.WithSuperRole(RoleOwner),
	)
	perms.SetPublicPermissions(
		PermissionThreadCreate,
		PermissionThreadRepost,
		PermissionReplyCreate,
	)
	perms.AddRole(
		RoleAdmin,
		PermissionBoardRename,
		PermissionBoardFlaggingUpdate,
		PermissionMemberInvite,
		PermissionMemberInviteRevoke,
		PermissionMemberRemove,
		PermissionThreadCreate,
		PermissionThreadEdit,
		PermissionThreadDelete,
		PermissionThreadRepost,
		PermissionThreadFlag,
		PermissionThreadFreeze,
		PermissionReplyCreate,
		PermissionReplyDelete,
		PermissionReplyFlag,
		PermissionRoleChange,
		PermissionUserBan,
		PermissionUserUnban,
	)
	perms.AddRole(
		RoleModerator,
		PermissionThreadCreate,
		PermissionThreadEdit,
		PermissionThreadRepost,
		PermissionThreadFlag,
		PermissionReplyCreate,
		PermissionReplyFlag,
		PermissionUserBan,
		PermissionUserUnban,
	)
	perms.AddRole(
		RoleGuest,
		PermissionThreadCreate,
		PermissionThreadRepost,
		PermissionReplyCreate,
	)
	perms.SetUserRoles(owner, RoleOwner)
	perms.ValidateFunc(PermissionBoardRename, validateOpenBoardRename)
	perms.ValidateFunc(PermissionMemberInvite, validateOpenMemberInvite)
	perms.ValidateFunc(PermissionRoleChange, validateOpenRoleChange)
	perms.ValidateFunc(PermissionThreadCreate, validateOpenThreadCreate)
	perms.ValidateFunc(PermissionReplyCreate, validateOpenReplyCreate)
	return perms
}
