package boards

type (
	// Permission defines the type for permissions.
	Permission string

	// Role defines the type for user roles.
	Role string

	// Args is a list of generic arguments.
	Args []interface{}

	// User contains user info.
	User struct {
		Address address
		Roles   []Role
	}

	// UsersIterFn defines a function type to iterate users.
	UsersIterFn func(User) bool

	// Permissions define an interface to for permissioned execution.
	Permissions interface {
		// HasRole checks if a user has a specific role assigned.
		HasRole(address, Role) bool

		// HasPermission checks if a user has a specific permission.
		HasPermission(address, Permission) bool

		// WithPermission calls a callback when a user has a specific permission.
		// Crossing when calling this method is required to allow implementations
		// to update their internal state, for example to create proposals that
		// when approved execute the permissioned callback.
		// It panics on error.
		//
		// An inline crossing function call can be used by the implementation if
		// crossing is required to update its internal state, for example to create
		// proposals that when approved execute the callback:
		//
		//  func(realm) {
		//    // Update internal realm state
		//    // ...
		//  }(cross)
		WithPermission(address, Permission, Args, func())

		// SetUserRoles adds a new user when it doesn't exist and sets its roles.
		// Method can also be called to change the roles of an existing user.
		// It panics on error.
		SetUserRoles(address, ...Role)

		// RemoveUser removes a user from the permissioner.
		// It panics on error.
		RemoveUser(address) (removed bool)

		// HasUser checks if a user exists.
		HasUser(address) bool

		// UsersCount returns the total number of users the permissioner contains.
		UsersCount() int

		// IterateUsers iterates permissions' users.
		IterateUsers(start, count int, fn UsersIterFn) bool
	}
)
