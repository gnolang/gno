package groups

import "gno.land/p/moul/addrset"

// Role is a named bucket of addresses with optional metadata.
//
// A Role is always owned by a parent Group; the only way to obtain a *Role
// is Group.AddRole or Group.GetRole. See the Group doc for the realm-
// boundary rules that govern passing *Role values around.
type Role struct {
	name    string
	members *addrset.Set
	meta    any
}

// newRole constructs a new empty role with the given name. Unexported: the
// only valid path to a *Role is via Group.AddRole, which registers it in
// the parent Group's role registry. A detached Role has no useful API.
func newRole(name string) *Role {
	return &Role{
		name:    name,
		members: &addrset.Set{},
	}
}

// Name returns the role's registry name.
func (r *Role) Name() string {
	return r.name
}

// Members returns a mutable reference to the role's member set; mutations
// through the returned pointer affect the role.
//
// SECURITY: the returned *addrset.Set is mutable. Do not expose it to
// untrusted callers — use Role.Readonly().Members() for a
// cross-realm-safe view.
func (r *Role) Members() *addrset.Set {
	return r.members
}

// Meta returns the role's metadata slot. See the package doc for the rule
// against storing mutable pointers in meta.
func (r *Role) Meta() any {
	return r.meta
}

// SetMeta sets the role's metadata slot. Passing nil clears it.
//
// SECURITY: do NOT store a pointer whose type has a mutator method (this
// includes common /p/ types like *addrset.Set or *avl.Tree) if untrusted
// realms may hold a Readonly() view of this Group. Meta() returns the value
// untainted, so a foreign reader can invoke that method and borrow rule #2
// commits the write under this (the allocating) realm's authority. A direct
// field write through the pointer is still blocked by the realm-ownership
// gate — the leak is specifically mutator-method dispatch. Prefer value types
// with no internal pointers. See the package doc.
func (r *Role) SetMeta(meta any) {
	r.meta = meta
}

// Readonly returns a read-only view of the role.
func (r *Role) Readonly() *ReadonlyRole {
	return &ReadonlyRole{role: r}
}
