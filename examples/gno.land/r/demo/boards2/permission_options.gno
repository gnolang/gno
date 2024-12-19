package boards2

import "std"

// DefaultPermissionsOption configures an DefaultPermissions.
type DefaultPermissionsOption func(*DefaultPermissions)

// WithSuperRole assigns a super role.
// A super role is one that have all permissions.
// These type of role doesn't need to be mapped to any permission.
func WithSuperRole(r Role) DefaultPermissionsOption {
	return func(dp *DefaultPermissions) {
		dp.superRole = r
	}
}

// WithUser adds a user to default permissions with optional assigned roles.
func WithUser(user std.Address, roles ...Role) DefaultPermissionsOption {
	return func(dp *DefaultPermissions) {
		// TODO: Should we enforce that users are members of the DAO? [dp.dao.IsMember(user)]
		dp.users.Set(user.String(), append([]Role(nil), roles...))
	}
}

// WithRole add a role to default permissions with one or more assigned permissions.
func WithRole(r Role, p Permission, extra ...Permission) DefaultPermissionsOption {
	return func(dp *DefaultPermissions) {
		dp.roles.Set(string(r), append([]Permission{p}, extra...))
	}
}
