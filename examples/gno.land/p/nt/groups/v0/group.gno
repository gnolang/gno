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
// Roles. The zero value is not usable; construct with NewGroup.
//
// # Security
//
// A Group, and the *Role values it hands out, are meant to be allocated and
// held by the consuming realm. Three rules apply at realm boundaries:
//
//  1. Do not ACCEPT a *Group or *Role from an external/untrusted caller —
//     subsequent mutations would route to the allocating (attacker)
//     realm's authority, and a poisoned Group could cause DoS or
//     unexpected state.
//
//  2. Do not RETURN a *Group or *Role from any method or function callable
//     by untrusted realms. Return *ReadonlyGroup or *ReadonlyRole instead.
//
//  3. Do not TRUST a *ReadonlyGroup or *ReadonlyRole received from an
//     untrusted caller — it is a live handle over the sender's data, not a
//     snapshot; the contents are attacker-controlled and can change between
//     reads.
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

// IterateRoles walks roles in lexicographic name order, starting at offset
// and visiting up to count roles. The callback receives a *ReadonlyRole —
// deliberately not a *Role, so that plumbing an untrusted callback into the
// iteration cannot escalate into role mutation under this realm's authority.
// To mutate, capture names during iteration and revisit via GetRole from a
// trusted context after iteration returns; registry mutation (AddRole,
// RemoveRole) mid-iteration can panic and abort the transaction. fn returns
// true to stop; IterateRoles returns true if stopped early.
func (g *Group) IterateRoles(offset, count int, fn func(*ReadonlyRole) bool) (stopped bool) {
	return g.roles.IterateByOffset(offset, count, func(_ string, value any) bool {
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
// Implementation note: dedup tracks seen addresses in an internal addrset,
// costing O(N) memory in the total membership.
func (g *Group) TotalSize() int {
	n := 0
	g.visitDistinct(func(address) bool {
		n++
		return false
	})
	return n
}

// IterateAll walks every distinct address across base + all roles,
// deduplicated. Order: base first (in addrset order), then roles in name
// order, skipping addresses already yielded. offset and count apply to the
// deduplicated output, not the pre-dedup items; a negative offset counts as
// zero. fn returns true to stop; IterateAll returns true if stopped early.
//
// The walk is live: do not mutate the group (base set, member sets, or the
// role registry) from within fn — registry mutation mid-iteration can panic
// and abort the transaction. Collect addresses first, mutate after
// IterateAll returns.
//
// Implementation note: dedup tracks seen addresses in an internal addrset
// (O(N) memory in the addresses scanned); scanning stops as soon as the
// requested window has been served. For paginating a large Group without
// dedup, use Iterate (base only) or IterateRoles.
func (g *Group) IterateAll(offset, count int, fn func(addr address) bool) (stopped bool) {
	if count <= 0 {
		return false
	}
	if offset < 0 {
		offset = 0
	}
	seen := 0
	g.visitDistinct(func(a address) bool {
		if seen < offset {
			seen++
			return false
		}
		if fn(a) {
			stopped = true
			return true
		}
		seen++
		return seen-offset >= count
	})
	return stopped
}

// visitDistinct walks the base set then every role (in name order), calling
// visit once per distinct address the first time it is seen. visit returns
// true to stop the walk early.
func (g *Group) visitDistinct(visit func(addr address) bool) {
	seen := &addrset.Set{}
	done := false
	record := func(a address) bool {
		if !seen.Add(a) { // Add returns false when already seen
			return false
		}
		done = visit(a)
		return done
	}
	g.base.IterateByOffset(0, g.base.Size(), record)
	if done {
		return
	}
	g.roles.IterateByOffset(0, g.roles.Size(), func(_ string, value any) bool {
		r := value.(*Role)
		r.members.IterateByOffset(0, r.members.Size(), record)
		return done
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
