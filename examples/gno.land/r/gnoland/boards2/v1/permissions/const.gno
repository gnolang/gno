package permissions

import "gno.land/p/gnoland/boards"

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
