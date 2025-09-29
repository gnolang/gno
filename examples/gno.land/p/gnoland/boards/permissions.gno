package boards

import "std"

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
		WithPermission(realm, std.Address, Permission, Args, func(realm))

		// SetUserRoles adds a new user when it doesn't exist and sets its roles.
		// Method can also be called to change the roles of an existing user.
		// It panics on error.
		SetUserRoles(realm, std.Address, ...Role)

		// RemoveUser removes a user from the permissioner.
		// It panics on error.
		RemoveUser(realm, std.Address) (removed bool)

		// HasUser checks if a user exists.
		HasUser(std.Address) bool

		// UsersCount returns the total number of users the permissioner contains.
		UsersCount() int

		// IterateUsers iterates permissions' users.
		IterateUsers(start, count int, fn UsersIterFn) bool
	}
)
