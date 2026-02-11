package boards2

import "gno.land/r/gnoland/boards2/v1/permissions"

func createBasicBoardPermissions(owner address) *permissions.Permissions {
	perms := permissions.New(
		permissions.UseSingleUserRole(),
		permissions.WithSuperRole(permissions.RoleOwner),
	)
	perms.AddRole(
		permissions.RoleAdmin,
		permissions.PermissionBoardRename,
		permissions.PermissionBoardFlaggingUpdate,
		permissions.PermissionMemberInvite,
		permissions.PermissionMemberInviteRevoke,
		permissions.PermissionMemberRemove,
		permissions.PermissionThreadCreate,
		permissions.PermissionThreadEdit,
		permissions.PermissionThreadDelete,
		permissions.PermissionThreadRepost,
		permissions.PermissionThreadFlag,
		permissions.PermissionThreadFreeze,
		permissions.PermissionReplyCreate,
		permissions.PermissionReplyDelete,
		permissions.PermissionReplyFlag,
		permissions.PermissionRoleChange,
		permissions.PermissionUserBan,
		permissions.PermissionUserUnban,
	)
	perms.AddRole(
		permissions.RoleModerator,
		permissions.PermissionThreadCreate,
		permissions.PermissionThreadEdit,
		permissions.PermissionThreadRepost,
		permissions.PermissionThreadFlag,
		permissions.PermissionReplyCreate,
		permissions.PermissionReplyFlag,
		permissions.PermissionUserBan,
		permissions.PermissionUserUnban,
	)
	perms.AddRole(
		permissions.RoleGuest,
		permissions.PermissionThreadCreate,
		permissions.PermissionThreadRepost,
		permissions.PermissionReplyCreate,
	)
	perms.SetUserRoles(cross, owner, permissions.RoleOwner)
	perms.ValidateFunc(permissions.PermissionBoardRename, validateBasicBoardRename)
	perms.ValidateFunc(permissions.PermissionMemberInvite, validateBasicMemberInvite)
	perms.ValidateFunc(permissions.PermissionRoleChange, validateBasicRoleChange)
	return perms
}

func createOpenBoardPermissions(owner address) *permissions.Permissions {
	perms := permissions.New(
		permissions.UseSingleUserRole(),
		permissions.WithSuperRole(permissions.RoleOwner),
	)
	perms.SetPublicPermissions(
		permissions.PermissionThreadCreate,
		permissions.PermissionReplyCreate,
	)
	perms.AddRole(
		permissions.RoleAdmin,
		permissions.PermissionBoardRename,
		permissions.PermissionBoardFlaggingUpdate,
		permissions.PermissionMemberInvite,
		permissions.PermissionMemberInviteRevoke,
		permissions.PermissionMemberRemove,
		permissions.PermissionThreadCreate,
		permissions.PermissionThreadEdit,
		permissions.PermissionThreadDelete,
		permissions.PermissionThreadRepost,
		permissions.PermissionThreadFlag,
		permissions.PermissionThreadFreeze,
		permissions.PermissionReplyCreate,
		permissions.PermissionReplyDelete,
		permissions.PermissionReplyFlag,
		permissions.PermissionRoleChange,
		permissions.PermissionUserBan,
		permissions.PermissionUserUnban,
	)
	perms.AddRole(
		permissions.RoleModerator,
		permissions.PermissionThreadCreate,
		permissions.PermissionThreadEdit,
		permissions.PermissionThreadRepost,
		permissions.PermissionThreadFlag,
		permissions.PermissionReplyCreate,
		permissions.PermissionReplyFlag,
		permissions.PermissionUserBan,
		permissions.PermissionUserUnban,
	)
	perms.AddRole(
		permissions.RoleGuest,
		permissions.PermissionThreadCreate,
		permissions.PermissionThreadRepost,
		permissions.PermissionReplyCreate,
	)
	perms.SetUserRoles(cross, owner, permissions.RoleOwner)
	perms.ValidateFunc(permissions.PermissionBoardRename, validateOpenBoardRename)
	perms.ValidateFunc(permissions.PermissionMemberInvite, validateOpenMemberInvite)
	perms.ValidateFunc(permissions.PermissionRoleChange, validateOpenRoleChange)
	perms.ValidateFunc(permissions.PermissionThreadCreate, validateOpenThreadCreate)
	perms.ValidateFunc(permissions.PermissionReplyCreate, validateOpenReplyCreate)
	return perms
}
