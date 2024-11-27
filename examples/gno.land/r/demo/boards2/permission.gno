package boards2

import "std"

const (
	PermissionBoardCreate  Permission = "board:create"
	PermissionThreadCreate            = "thread:create"
	PermissionThreadEdit              = "thread:edit"
	PermissionThreadDelete            = "thread:delete"
	PermissionThreadRepost            = "thread:repost"
	PermissionReplyDelete             = "reply:delete"
	PermissionMemberInvite            = "member:invite"
)

const (
	RoleOwner     Role = "owner"
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

	// Permissioner define an interface to for permissioned execution.
	Permissioner interface {
		// HasPermission checks if a user has a specific permission.
		HasPermission(std.Address, Permission) bool

		// WithPermission calls a callback when a user has a specific permission.
		// It panics on error.
		WithPermission(std.Address, Permission, Args, func(Args))

		// AddUser adds a new user to the permissioner.
		AddUser(std.Address, ...Role) error
	}
)
