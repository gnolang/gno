package boards2

import (
	"std"

	"gno.land/p/demo/boards2/admindao"
)

const (
	PermissionBoardCreate  Permission = "board:create"
	PermissionBoardRename             = "board:rename"
	PermissionThreadCreate            = "thread:create"
	PermissionThreadEdit              = "thread:edit"
	PermissionThreadDelete            = "thread:delete"
	PermissionThreadFlag              = "thread:flag"
	PermissionThreadRepost            = "thread:repost"
	PermissionReplyDelete             = "reply:delete"
	PermissionReplyFlag               = "reply:flag"
	PermissionMemberInvite            = "member:invite"
	PermissionMemberRemove            = "member:remove"
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

		// GetDAO returns the underlying DAO.
		// Returned value can be nil if the implementation doesn't have a DAO.
		GetDAO() *admindao.AdminDAO // TODO: should return an interface
	}
)
