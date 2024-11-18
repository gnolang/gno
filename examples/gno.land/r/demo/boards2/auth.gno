package boards

import (
	"errors"
	"std"
)

const (
	PermissionEdit   Permission = "edit"
	PermissionDelete Permission = "delete"
)

// ErrUnauzorized indicates that user doesn't have a required permission.
var ErrUnauzorized = errors.New("unauthorized")

type (
	// Permission defines the type for permissions.
	Permission string

	// Role defines the type for user roles.
	Role string

	// Args is a list of generic arguments.
	Args []interface{}

	// Auth define an interface to for permissioned execution.
	Auth interface {
		// WithPermission calls a callback when a user has a specific permission.
		WithPermission(std.Address, Permission, Args, func(Args)) error

		// GetRoles returns the list of roles.
		GetRoles() []Role

		// GetUsers returns the list of users that have a specific role.
		GetUsers(Role) []std.Address
	}
)
