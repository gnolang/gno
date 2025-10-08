package boards2

import (
	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/commondao"
)

const (
	RoleGuest     boards.Role = ""
	RoleOwner                 = "owner"
	RoleAdmin                 = "admin"
	RoleModerator             = "moderator"
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
	PermissionReplyFreeze                           = "reply:freeze"
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

func createBasicBoardPermissions(owner address) *BasicPermissions {
	dao := commondao.New(commondao.WithMember(owner))
	perms := NewBasicPermissions(dao)
	perms.SetSuperRole(RoleOwner)
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
		PermissionReplyFreeze,
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
	perms.SetUserRoles(cross, owner, RoleOwner)
	perms.ValidateFunc(PermissionBoardRename, validateBoardRename)
	perms.ValidateFunc(PermissionMemberInvite, validateMemberInvite)
	perms.ValidateFunc(PermissionRoleChange, validateRoleChange)
	return perms
}
