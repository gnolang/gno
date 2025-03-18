package boards2

import "std"

const (
	PermissionRealmNotice         Permission = "realm:notice"
	PermissionBoardCreate                    = "board:create"
	PermissionBoardRename                    = "board:rename"
	PermissionBoardFlaggingUpdate            = "board:flagging-update"
	PermissionBoardFreeze                    = "board:freeze"
	PermissionThreadCreate                   = "thread:create"
	PermissionThreadEdit                     = "thread:edit"
	PermissionThreadDelete                   = "thread:delete"
	PermissionThreadFreeze                   = "thread:freeze"
	PermissionThreadFlag                     = "thread:flag"
	PermissionThreadRepost                   = "thread:repost"
	PermissionReplyCreate                    = "reply:create"
	PermissionReplyDelete                    = "reply:delete"
	PermissionReplyFlag                      = "reply:flag"
	PermissionReplyFreeze                    = "reply:freeze"
	PermissionMemberInvite                   = "member:invite"
	PermissionMemberRemove                   = "member:remove"
	PermissionRoleChange                     = "role:change"
	PermissionPermissionsUpdate              = "permissions:update"
)

const (
	RoleOwner     Role = "owner"
	RoleAdmin          = "admin"
	RoleModerator      = "moderator"
	RoleGuest          = ""
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
	Permissions interface {
		// HasRole checks if a user has a specific role assigned.
		HasRole(std.Address, Role) bool

		// HasPermission checks if a user has a specific permission.
		HasPermission(std.Address, Permission) bool

		// WithPermission calls a callback when a user has a specific permission.
		// It panics on error.
		WithPermission(std.Address, Permission, Args, func(Args))

		// AddUser adds a new user to the permissioner.
		AddUser(std.Address, ...Role) error

		// SetUserRoles sets the roles of a user.
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
