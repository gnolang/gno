package groups

import (
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	addr0 = address("addr0")
	addr1 = address("addr1")
	addr2 = address("addr2")
	addr3 = address("addr3")
	addr4 = address("addr4")
	addr5 = address("addr5")
	addr6 = address("addr6")
)

// newTestGroup builds the fixture used across tests:
//
//	base:   addr0, addr1, addr3, addr5
//	admins: addr2, addr3, addr5
//	mods:   addr1, addr4, addr5
//
// Overlap topology: addr0 base-only, addr2 admins-only, addr4 mods-only,
// addr1 base+mods, addr3 base+admins, addr5 base+both roles.
// Distinct addresses: addr0..addr5 (addr6 is never a member).
func newTestGroup(t *testing.T) *Group {
	t.Helper()
	g := NewGroup()
	g.Add(addr0)
	g.Add(addr1)
	g.Add(addr3)
	g.Add(addr5)

	admins, err := g.AddRole("admins")
	urequire.NoError(t, err)
	admins.Members().Add(addr2)
	admins.Members().Add(addr3)
	admins.Members().Add(addr5)

	mods, err := g.AddRole("mods")
	urequire.NoError(t, err)
	mods.Members().Add(addr1)
	mods.Members().Add(addr4)
	mods.Members().Add(addr5)
	return g
}

func joinAddrs(addrs []address) string {
	strs := make([]string, 0, len(addrs))
	for _, a := range addrs {
		strs = append(strs, string(a))
	}
	return strings.Join(strs, ",")
}

func TestBaseSet(t *testing.T) {
	g := NewGroup()
	uassert.Equal(t, 0, g.Size())
	uassert.False(t, g.Has(addr1))

	uassert.True(t, g.Add(addr1))
	uassert.False(t, g.Add(addr1)) // duplicate
	uassert.True(t, g.Add(addr2))
	uassert.Equal(t, 2, g.Size())
	uassert.True(t, g.Has(addr1))
	uassert.False(t, g.Has(addr3))

	uassert.True(t, g.Remove(addr1))
	uassert.False(t, g.Remove(addr1)) // already gone
	uassert.Equal(t, 1, g.Size())
	uassert.False(t, g.Has(addr1))
}

func TestIterate(t *testing.T) {
	g := NewGroup()
	g.Add(addr2)
	g.Add(addr1)
	g.Add(addr3)

	var got []address
	stopped := g.Iterate(0, 10, func(a address) bool {
		got = append(got, a)
		return false
	})
	uassert.False(t, stopped)
	uassert.Equal(t, "addr1,addr2,addr3", joinAddrs(got)) // sorted order

	got = nil
	g.Iterate(1, 1, func(a address) bool {
		got = append(got, a)
		return false
	})
	uassert.Equal(t, "addr2", joinAddrs(got))

	// Early stop halts iteration, not just the return value.
	visited := 0
	stopped = g.Iterate(0, 10, func(a address) bool {
		visited++
		return true
	})
	uassert.True(t, stopped)
	uassert.Equal(t, 1, visited)

	// Roles are never consulted.
	gg := newTestGroup(t)
	got = nil
	gg.Iterate(0, 10, func(a address) bool {
		got = append(got, a)
		return false
	})
	uassert.Equal(t, "addr0,addr1,addr3,addr5", joinAddrs(got))
}

func TestAddRole(t *testing.T) {
	g := NewGroup()

	r, err := g.AddRole("admins")
	urequire.NoError(t, err)
	uassert.Equal(t, "admins", r.Name())
	uassert.Equal(t, 0, r.Members().Size())
	uassert.Equal(t, 1, g.RoleCount())

	_, err = g.AddRole("admins")
	uassert.ErrorIs(t, err, ErrRoleExists)

	_, err = g.AddRole("")
	uassert.ErrorIs(t, err, ErrEmptyName)

	uassert.Equal(t, 1, g.RoleCount())
}

func TestGetRoleHasRole(t *testing.T) {
	g := newTestGroup(t)

	r, found := g.GetRole("admins")
	uassert.True(t, found)
	uassert.Equal(t, "admins", r.Name())
	uassert.True(t, r.Members().Has(addr2))

	_, found = g.GetRole("nope")
	uassert.False(t, found)

	uassert.True(t, g.HasRole("mods"))
	uassert.False(t, g.HasRole("nope"))
}

func TestRemoveRole(t *testing.T) {
	g := newTestGroup(t)

	uassert.True(t, g.RemoveRole("admins"))
	uassert.False(t, g.RemoveRole("admins")) // already gone
	uassert.Equal(t, 1, g.RoleCount())
	uassert.False(t, g.HasRole("admins"))

	// Members of the removed role stay in the base set and other roles.
	uassert.True(t, g.Has(addr3))     // base
	uassert.True(t, g.HasAny(addr5))  // still in base + mods
	uassert.False(t, g.HasAny(addr2)) // was only in admins
}

func TestIterateRoles(t *testing.T) {
	g := newTestGroup(t)

	var names []string
	stopped := g.IterateRoles(0, 10, func(r *Role) bool {
		names = append(names, r.Name())
		return false
	})
	uassert.False(t, stopped)
	uassert.Equal(t, "admins,mods", strings.Join(names, ","))

	// offset/count window.
	names = nil
	g.IterateRoles(1, 1, func(r *Role) bool {
		names = append(names, r.Name())
		return false
	})
	uassert.Equal(t, "mods", strings.Join(names, ","))

	// Early stop halts iteration, not just the return value.
	names = nil
	stopped = g.IterateRoles(0, 10, func(r *Role) bool {
		names = append(names, r.Name())
		return true
	})
	uassert.True(t, stopped)
	uassert.Equal(t, "admins", strings.Join(names, ","))

	// Yielded roles are mutable; member-set mutation during iteration is
	// supported (registry mutation is not).
	g.IterateRoles(0, 10, func(r *Role) bool {
		r.Members().Add(addr6)
		return false
	})
	uassert.Equal(t, "admins,mods", strings.Join(g.RolesContaining(addr6), ","))
}

func TestRoleMeta(t *testing.T) {
	g := NewGroup()
	r, err := g.AddRole("admins")
	urequire.NoError(t, err)

	uassert.Nil(t, r.Meta())
	r.SetMeta("quorum=2/3")
	uassert.Equal(t, "quorum=2/3", r.Meta())
	uassert.Equal(t, "quorum=2/3", r.Readonly().Meta())
	r.SetMeta(nil)
	uassert.Nil(t, r.Meta())
}

func TestHasAny(t *testing.T) {
	g := newTestGroup(t)

	uassert.True(t, g.HasAny(addr0))  // base only
	uassert.True(t, g.HasAny(addr1))  // base + mods
	uassert.True(t, g.HasAny(addr2))  // admins only
	uassert.True(t, g.HasAny(addr4))  // mods only
	uassert.True(t, g.HasAny(addr5))  // base + both roles
	uassert.False(t, g.HasAny(addr6)) // nowhere

	// Base-only member, no roles registered at all.
	g2 := NewGroup()
	g2.Add(addr1)
	uassert.True(t, g2.HasAny(addr1))
	uassert.False(t, g2.HasAny(addr2))
}

func TestTotalSize(t *testing.T) {
	uassert.Equal(t, 0, NewGroup().TotalSize())

	g := newTestGroup(t)
	uassert.Equal(t, 4, g.Size())      // base only
	uassert.Equal(t, 6, g.TotalSize()) // addr0..addr5, deduplicated
}

func TestIterateAll(t *testing.T) {
	g := newTestGroup(t)

	collect := func(offset, count int) (addrs []address, stopped bool) {
		stopped = g.IterateAll(offset, count, func(a address) bool {
			addrs = append(addrs, a)
			return false
		})
		return addrs, stopped
	}

	// Order: base first (sorted), then roles in name order, deduplicated.
	got, stopped := collect(0, 10)
	uassert.False(t, stopped)
	uassert.Equal(t, "addr0,addr1,addr3,addr5,addr2,addr4", joinAddrs(got))

	// offset/count window over the deduplicated output.
	got, _ = collect(1, 2)
	uassert.Equal(t, "addr1,addr3", joinAddrs(got))
	got, _ = collect(3, 10)
	uassert.Equal(t, "addr5,addr2,addr4", joinAddrs(got))
	// Window straddling skipped duplicates: count applies to deduplicated
	// output, not to scanned items.
	got, _ = collect(4, 2)
	uassert.Equal(t, "addr2,addr4", joinAddrs(got))

	// Degenerate parameters.
	got, stopped = collect(6, 10) // offset == total
	uassert.False(t, stopped)
	uassert.Equal(t, 0, len(got))
	got, _ = collect(0, 0)
	uassert.Equal(t, 0, len(got))
	got, _ = collect(-1, 2) // negative offset counts as zero
	uassert.Equal(t, "addr0,addr1", joinAddrs(got))

	// Early stop halts iteration, not just the return value.
	visited := 0
	stopped = g.IterateAll(0, 10, func(a address) bool {
		visited++
		return true
	})
	uassert.True(t, stopped)
	uassert.Equal(t, 1, visited)
}

func TestRolesContaining(t *testing.T) {
	g := newTestGroup(t)

	uassert.Equal(t, "admins,mods", strings.Join(g.RolesContaining(addr5), ","))
	uassert.Equal(t, "admins", strings.Join(g.RolesContaining(addr2), ","))
	uassert.Equal(t, "mods", strings.Join(g.RolesContaining(addr1), ",")) // base not consulted
	uassert.Equal(t, 0, len(g.RolesContaining(addr0)))                    // base-only member
	uassert.Equal(t, 0, len(g.RolesContaining(addr6)))
}

func TestRemoveFromAll(t *testing.T) {
	g := newTestGroup(t)

	uassert.True(t, g.RemoveFromAll(addr5)) // base + both roles
	uassert.False(t, g.HasAny(addr5))

	uassert.True(t, g.RemoveFromAll(addr0)) // base-only: removed reflects the base hit alone
	uassert.False(t, g.HasAny(addr0))

	uassert.True(t, g.RemoveFromAll(addr2)) // role-only
	uassert.False(t, g.HasAny(addr2))

	uassert.False(t, g.RemoveFromAll(addr6)) // nowhere

	uassert.Equal(t, 3, g.TotalSize()) // addr1, addr3, addr4 remain
	uassert.Equal(t, 2, g.RoleCount()) // roles themselves are kept
}

func TestReadonlyGroup(t *testing.T) {
	g := newTestGroup(t)
	rg := g.Readonly()

	// Base reads.
	uassert.True(t, rg.Has(addr0))
	uassert.False(t, rg.Has(addr2))
	uassert.Equal(t, 4, rg.Size())
	var got []address
	stopped := rg.Iterate(0, 10, func(a address) bool {
		got = append(got, a)
		return false
	})
	uassert.False(t, stopped)
	uassert.Equal(t, "addr0,addr1,addr3,addr5", joinAddrs(got))

	// Role registry reads.
	uassert.True(t, rg.HasRole("admins"))
	uassert.False(t, rg.HasRole("nope"))
	uassert.Equal(t, 2, rg.RoleCount())
	var names []string
	stopped = rg.IterateRoles(0, 10, func(rr *ReadonlyRole) bool {
		names = append(names, rr.Name())
		return false
	})
	uassert.False(t, stopped)
	uassert.Equal(t, "admins,mods", strings.Join(names, ","))

	rr, found := rg.GetRole("admins")
	uassert.True(t, found)
	uassert.Equal(t, "admins", rr.Name())
	uassert.True(t, rr.Members().Has(addr2))
	uassert.Equal(t, 3, rr.Members().Size())
	_, found = rg.GetRole("nope")
	uassert.False(t, found)

	// Aggregated reads.
	uassert.True(t, rg.HasAny(addr4))
	uassert.False(t, rg.HasAny(addr6))
	uassert.Equal(t, 6, rg.TotalSize())
	got = nil
	stopped = rg.IterateAll(0, 10, func(a address) bool {
		got = append(got, a)
		return false
	})
	uassert.False(t, stopped)
	uassert.Equal(t, "addr0,addr1,addr3,addr5,addr2,addr4", joinAddrs(got))
	uassert.True(t, rg.IterateAll(0, 10, func(a address) bool { return true }))
	uassert.Equal(t, "admins,mods", strings.Join(rg.RolesContaining(addr5), ","))
}

func TestReadonlyViewsAreLive(t *testing.T) {
	g := newTestGroup(t)
	rg := g.Readonly()
	admins, found := g.GetRole("admins")
	urequire.True(t, found)
	rr := admins.Readonly()

	// Views are thin handles, not snapshots: mutations through the owning
	// side are immediately visible through previously created views.
	g.Add(addr6)
	uassert.Equal(t, 5, rg.Size())
	uassert.True(t, rg.Has(addr6))

	admins.Members().Add(addr6)
	uassert.Equal(t, 4, rr.Members().Size())
	uassert.True(t, rr.Members().Has(addr6))
}
