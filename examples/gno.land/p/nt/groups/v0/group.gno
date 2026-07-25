package groups

import (
	"errors"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/bptree/v0"
)

var (
	ErrRoleExists = errors.New("role already exists")
	ErrEmptyName  = errors.New("role name is required")
)

// Group is a container with a base address set plus a registry of named
// Roles.
//
// # Security
//
// A Group, and the *Role values it hands out, are meant to be allocated and
// held by the consuming realm. Two rules apply at realm boundaries:
//
//  1. Do not ACCEPT a *Group or *Role from an external/untrusted caller —
//     subsequent mutations would route to the allocating (attacker)
//     realm's authority, and a poisoned Group could cause DoS or
//     unexpected state.
//
//  2. Do not RETURN a *Group or *Role from any method or function callable
//     by untrusted realms. Return *ReadonlyGroup or *ReadonlyRole instead.
//
// Both directions matter: exposing a *Group to attacker code is as
// dangerous as accepting one. The Readonly() views are the only safe
// handles to cross a realm boundary.
type Group struct {
	base  *addrset.Set
	roles *bptree.BPTree // name -> *Role
}

// NewGroup constructs an empty group.
func NewGroup() *Group {
	return &Group{
		base:  &addrset.Set{},
		roles: bptree.NewBPTree32(),
	}
}

// --- Base set ---
//
// All base methods operate ONLY on the base set; roles are never consulted.
// Use the aggregated forms (HasAny, TotalSize, IterateAll, RemoveFromAll)
// for views across base + roles.

// Add inserts addr into the base set. Returns true if newly added.
func (g *Group) Add(addr address) (added bool) {
	return g.base.Add(addr)
}

// Remove deletes addr from the base set. Returns true if it was present.
func (g *Group) Remove(addr address) (removed bool) {
	return g.base.Remove(addr)
}

// Has reports whether addr is in the base set. It does NOT consult roles;
// use HasAny for an aggregated check.
func (g *Group) Has(addr address) bool {
	return g.base.Has(addr)
}

// Size returns the number of addresses in the base set only.
func (g *Group) Size() int {
	return g.base.Size()
}

// Iterate walks the base set (only) in sorted order, starting at offset.
// fn returns true to stop; Iterate returns true if stopped early.
func (g *Group) Iterate(offset, count int, fn func(addr address) bool) (stopped bool) {
	g.base.IterateByOffset(offset, count, func(a address) bool {
		stopped = fn(a)
		return stopped
	})
	return stopped
}

// --- Role registry ---

// AddRole registers a new empty role. Returns ErrEmptyName if name is
// empty, ErrRoleExists if a role of that name already exists.
func (g *Group) AddRole(name string) (*Role, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	if g.roles.Has(name) {
		return nil, ErrRoleExists
	}
	r := newRole(name)
	g.roles.Set(name, r)
	return r, nil
}

// GetRole returns the mutable role if it exists.
//
// SECURITY: the returned *Role exposes mutators (Members().Add/Remove,
// SetMeta). Do not pass it to untrusted callers — use GetRole on a
// *ReadonlyGroup for cross-realm exposure.
func (g *Group) GetRole(name string) (r *Role, found bool) {
	r, found = g.roles.Get(name).(*Role)
	return r, found
}

// HasRole reports whether a role with the given name exists.
func (g *Group) HasRole(name string) bool {
	return g.roles.Has(name)
}

// RemoveRole removes the named role and its membership records. Members of
// the removed role are NOT removed from the base set or from any other
// role; only this role's own data is discarded. Returns false if no such
// role exists.
func (g *Group) RemoveRole(name string) (removed bool) {
	_, removed = g.roles.Remove(name)
	return removed
}

// RoleCount returns the number of registered roles.
func (g *Group) RoleCount() int {
	return g.roles.Size()
}

// IterateRoles walks every role in lexicographic name order. The callback
// receives a *ReadonlyRole — mutation during iteration is not supported;
// capture the name and revisit via GetRole from a trusted context. fn
// returns true to stop; IterateRoles returns true if stopped early.
func (g *Group) IterateRoles(fn func(*ReadonlyRole) bool) (stopped bool) {
	return g.roles.IterateByOffset(0, g.roles.Size(), func(_ string, value any) bool {
		return fn(value.(*Role).Readonly())
	})
}

// --- Aggregations across base + all roles ---

// HasAny reports whether addr is in the base set OR in any role.
func (g *Group) HasAny(addr address) bool {
	if g.base.Has(addr) {
		return true
	}
	found := false
	g.roles.IterateByOffset(0, g.roles.Size(), func(_ string, value any) bool {
		if value.(*Role).members.Has(addr) {
			found = true
			return true // stop
		}
		return false
	})
	return found
}

// TotalSize returns the count of distinct addresses across the base set and
// all roles, deduplicated. A caller may place the same address in base and
// in multiple roles; TotalSize counts it once.
//
// Implementation note: builds an internal addrset to dedup, costing O(N)
// memory in the total membership.
func (g *Group) TotalSize() int {
	seen := &addrset.Set{}
	g.collectDistinct(seen, nil)
	return seen.Size()
}

// IterateAll walks every distinct address across base + all roles,
// deduplicated. Order: base first (in addrset order), then roles in name
// order, skipping addresses already yielded. offset and count apply to the
// final deduplicated output, not the pre-dedup items. fn returns true to
// stop; IterateAll returns true if stopped early.
//
// Implementation note: builds the full deduplicated view upfront (O(N)
// memory) even for small count. For paginating a large Group without dedup,
// use Iterate (base only) or IterateRoles + GetRole.
func (g *Group) IterateAll(offset, count int, fn func(addr address) bool) (stopped bool) {
	if count <= 0 || offset < 0 {
		return false
	}
	var ordered []address
	g.collectDistinct(&addrset.Set{}, func(a address) {
		ordered = append(ordered, a)
	})
	if offset >= len(ordered) {
		return false
	}
	end := offset + count
	if end > len(ordered) {
		end = len(ordered)
	}
	for _, a := range ordered[offset:end] {
		if fn(a) {
			return true
		}
	}
	return false
}

// collectDistinct walks base then roles (name order), recording each
// address into seen. emit, if non-nil, is called once per distinct address
// in iteration order (the first time it is seen).
func (g *Group) collectDistinct(seen *addrset.Set, emit func(addr address)) {
	visit := func(a address) bool {
		if seen.Add(a) && emit != nil { // Add returns true when newly added
			emit(a)
		}
		return false
	}
	g.base.IterateByOffset(0, g.base.Size(), visit)
	g.roles.IterateByOffset(0, g.roles.Size(), func(_ string, value any) bool {
		r := value.(*Role)
		r.members.IterateByOffset(0, r.members.Size(), visit)
		return false
	})
}

// RolesContaining returns the names of all roles containing addr, in
// lexicographic name order. The base set is not consulted (base membership
// is not a "role"). Returns nil if addr is in no roles.
func (g *Group) RolesContaining(addr address) []string {
	var names []string
	g.roles.IterateByOffset(0, g.roles.Size(), func(name string, value any) bool {
		if value.(*Role).members.Has(addr) {
			names = append(names, name)
		}
		return false
	})
	return names
}

// RemoveFromAll removes addr from the base set and from every role. Returns
// true if it was removed from at least one location.
func (g *Group) RemoveFromAll(addr address) (removed bool) {
	if g.base.Remove(addr) {
		removed = true
	}
	g.roles.IterateByOffset(0, g.roles.Size(), func(_ string, value any) bool {
		if value.(*Role).members.Remove(addr) {
			removed = true
		}
		return false
	})
	return removed
}

// Readonly returns a read-only view of the group.
func (g *Group) Readonly() *ReadonlyGroup {
	return &ReadonlyGroup{group: g}
}
