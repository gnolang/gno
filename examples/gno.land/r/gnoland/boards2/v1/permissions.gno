package boards2

import (
	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/permissions"
)

const (
	RoleOwner     boards.Role = "owner"
	RoleAdmin                 = "admin"
	RoleModerator             = "moderator"
	RoleGuest                 = "guest"
)

const (
	PermissionBoardCreate         boards.Permission = "board:create"
	PermissionBoardFlaggingUpdate                   = "board:flagging-update"
	PermissionBoardFreeze                           = "board:freeze"
	PermissionBoardRename                           = "board:rename"
	PermissionMemberInvite                          = "member:invite"
	PermissionMemberInviteRevoke                    = "member:invite-remove"
	PermissionMemberRemove                          = "member:remove"
	PermissionPermissionsUpdate                     = "permissions:update"
	PermissionRealmHelp                             = "realm:help"
	PermissionRealmLock                             = "realm:lock"
	PermissionRealmNotice                           = "realm:notice"
	PermissionReplyCreate                           = "reply:create"
	PermissionReplyDelete                           = "reply:delete"
	PermissionReplyFlag                             = "reply:flag"
	PermissionRoleChange                            = "role:change"
	PermissionThreadCreate                          = "thread:create"
	PermissionThreadDelete                          = "thread:delete"
	PermissionThreadEdit                            = "thread:edit"
	PermissionThreadFlag                            = "thread:flag"
	PermissionThreadFreeze                          = "thread:freeze"
	PermissionThreadRepost                          = "thread:repost"
	PermissionUserBan                               = "user:ban"
	PermissionUserUnban                             = "user:unban"
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
