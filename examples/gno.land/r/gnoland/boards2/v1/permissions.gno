package boards2

import "std"

const (
	PermissionBoardCreate         Permission = "board:create"
	PermissionBoardFlaggingUpdate            = "board:flagging-update"
	PermissionBoardFreeze                    = "board:freeze"
	PermissionBoardRename                    = "board:rename"
	PermissionMemberInvite                   = "member:invite"
	PermissionMemberInviteRevoke             = "member:invite-remove"
	PermissionMemberRemove                   = "member:remove"
	PermissionPermissionsUpdate              = "permissions:update"
	PermissionRealmHelp                      = "realm:help"
	PermissionRealmLock                      = "realm:lock"
	PermissionRealmNotice                    = "realm:notice"
	PermissionReplyCreate                    = "reply:create"
	PermissionReplyDelete                    = "reply:delete"
	PermissionReplyFlag                      = "reply:flag"
	PermissionReplyFreeze                    = "reply:freeze"
	PermissionRoleChange                     = "role:change"
	PermissionThreadCreate                   = "thread:create"
	PermissionThreadDelete                   = "thread:delete"
	PermissionThreadEdit                     = "thread:edit"
	PermissionThreadFlag                     = "thread:flag"
	PermissionThreadFreeze                   = "thread:freeze"
	PermissionThreadRepost                   = "thread:repost"
	PermissionUserBan                        = "user:ban"
	PermissionUserUnban                      = "user:unban"
)

const (
	RoleGuest     Role = ""
	RoleOwner          = "owner"
	RoleAdmin          = "admin"
	RoleModerator      = "moderator"
)

type (
	// Permission defines the type for permissions.
	Permission string

	// Role defines the type for user roles.
	Role string

	// Args is a list of generic arguments.
	Args []interface{}

	// User contains user info.
	User struct {
		Address std.Address
		Roles   []Role
	}

	// UsersIterFn defines a function type to iterate users.
	UsersIterFn func(User) bool

	// Permissions define an interface to for permissioned execution.
	// TODO: Add crossing support to Permissions
	Permissions interface {
		// HasRole checks if a user has a specific role assigned.
		HasRole(std.Address, Role) bool

		// HasPermission checks if a user has a specific permission.
		HasPermission(std.Address, Permission) bool

		// WithPermission calls a callback when a user has a specific permission.
		// It panics on error.
		WithPermission(std.Address, Permission, Args, func(Args))

		// SetUserRoles adds a new user when it doesn't exist and sets its roles.
		// Method can also be called to change the roles of an existing user.
		SetUserRoles(std.Address, ...Role) error

		// RemoveUser removes a user from the permissioner.
		RemoveUser(std.Address) (removed bool)

		// HasUser checks if a user exists.
		HasUser(std.Address) bool

		// UsersCount returns the total number of users the permissioner contains.
		UsersCount() int

		// IterateUsers iterates permissions' users.
		IterateUsers(start, count int, fn UsersIterFn) bool
	}
)
