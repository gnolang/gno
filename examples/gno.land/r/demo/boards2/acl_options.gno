package boards

import "std"

// ACLOption configures an ACL.
type ACLOption func(*ACL)

// WithSuperRole assigns a super role.
// A super role is one that have all ACL permissions.
// These type of role doesn't need to be mapped to any permission.
func WithSuperRole(r Role) ACLOption {
	return func(acl *ACL) {
		acl.superRole = r
	}
}

// WithUser adds a user to the ACL with optional assigned roles.
func WithUser(user std.Address, roles ...Role) ACLOption {
	return func(acl *ACL) {
		// TODO: Should we enforce that users are members of the DAO? [acl.dao.IsMember(user)]
		acl.users.Set(user.String(), append([]Role(nil), roles...))
	}
}

// WithRole add a role to the ACL with one or more assigned permissions.
func WithRole(r Role, p Permission, extra ...Permission) ACLOption {
	return func(acl *ACL) {
		acl.roles.Set(string(r), append([]Permission{p}, extra...))
	}
}
