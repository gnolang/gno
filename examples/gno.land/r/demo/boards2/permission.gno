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
	PermissionMemberRemove            = "member:remove"
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

		// RemoveUser removes a user from the permissioner.
		RemoveUser(std.Address) (removed bool)
	}
)
