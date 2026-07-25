package groups

import "gno.land/p/moul/addrset"

// ReadonlyRole is a read-only view of a Role. It exposes only read-side
// methods and holds the *Role in an unexported field, so cross-package
// callers cannot mutate the role through this type.
type ReadonlyRole struct {
	role *Role
}

// Name returns the role's name.
func (rr ReadonlyRole) Name() string {
	return rr.role.name
}

// Members returns a read-only view of the role's member set.
func (rr ReadonlyRole) Members() *addrset.ReadonlySet {
	return rr.role.members.Readonly()
}

// Meta returns the role's metadata slot.
//
// NOTE: a mutable pointer stored in meta is NOT protected by this readonly
// view — the pointee remains mutable by anyone who retrieves it. See the
// package doc.
func (rr ReadonlyRole) Meta() any {
	return rr.role.meta
}

// ReadonlyGroup is a read-only view of a Group. Every method mirrors the
// read-side of Group; mutators are absent. It holds the *Group in an
// unexported field, so cross-package callers cannot mutate through it.
type ReadonlyGroup struct {
	group *Group
}

// --- Base set (base only) ---

// Has reports whether addr is in the base set (roles not consulted).
func (rg ReadonlyGroup) Has(addr address) bool {
	return rg.group.Has(addr)
}

// Size returns the number of addresses in the base set only.
func (rg ReadonlyGroup) Size() int {
	return rg.group.Size()
}

// Iterate walks the base set (only); see Group.Iterate.
func (rg ReadonlyGroup) Iterate(offset, count int, fn func(addr address) bool) (stopped bool) {
	return rg.group.Iterate(offset, count, fn)
}

// --- Role registry ---

// GetRole returns a read-only view of the named role if it exists.
func (rg ReadonlyGroup) GetRole(name string) (rr *ReadonlyRole, found bool) {
	r, ok := rg.group.GetRole(name)
	if !ok {
		return nil, false
	}
	return r.Readonly(), true
}

// HasRole reports whether a role with the given name exists.
func (rg ReadonlyGroup) HasRole(name string) bool {
	return rg.group.HasRole(name)
}

// RoleCount returns the number of registered roles.
func (rg ReadonlyGroup) RoleCount() int {
	return rg.group.RoleCount()
}

// IterateRoles walks roles in name order as read-only views; see
// Group.IterateRoles.
func (rg ReadonlyGroup) IterateRoles(offset, count int, fn func(*ReadonlyRole) bool) (stopped bool) {
	return rg.group.IterateRoles(offset, count, func(r *Role) bool {
		return fn(r.Readonly())
	})
}

// --- Aggregations ---

// HasAny reports whether addr is in the base set OR in any role.
func (rg ReadonlyGroup) HasAny(addr address) bool {
	return rg.group.HasAny(addr)
}

// TotalSize returns the deduplicated count across base + all roles.
func (rg ReadonlyGroup) TotalSize() int {
	return rg.group.TotalSize()
}

// IterateAll walks every distinct address across base + all roles; see
// Group.IterateAll.
func (rg ReadonlyGroup) IterateAll(offset, count int, fn func(addr address) bool) (stopped bool) {
	return rg.group.IterateAll(offset, count, fn)
}

// RolesContaining returns the names of all roles containing addr, in name
// order; see Group.RolesContaining.
func (rg ReadonlyGroup) RolesContaining(addr address) []string {
	return rg.group.RolesContaining(addr)
}
