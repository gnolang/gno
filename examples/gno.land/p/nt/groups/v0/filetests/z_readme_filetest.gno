// End-to-end scenario from the package README: a group with a base set and
// three roles, exercising base vs aggregated semantics, dedup, role
// removal, and readonly views.
package main

import (
	"strings"

	"gno.land/p/nt/groups/v0"
)

var (
	alice = address("alice") // base only
	bob   = address("bob")   // base + council
	carol = address("carol") // base + admin
	dave  = address("dave")  // council only
	eve   = address("eve")   // admin + moderator
	frank = address("frank") // moderator only
	grace = address("grace") // admin only
	zed   = address("zed")   // not a member
)

func main() {
	// 1. A group with a base set and three named roles.
	g := groups.NewGroup()
	g.Add(alice)
	g.Add(bob)
	g.Add(carol)

	admin, _ := g.AddRole("admin")
	admin.Members().Add(carol)
	admin.Members().Add(eve)
	admin.Members().Add(grace)

	council, _ := g.AddRole("council")
	council.Members().Add(bob)
	council.Members().Add(dave)

	moderator, _ := g.AddRole("moderator")
	moderator.Members().Add(eve)
	moderator.Members().Add(frank)

	// 2. Has (base only) vs HasAny (base + roles).
	println("Has(alice):", g.Has(alice))     // base member
	println("Has(dave):", g.Has(dave))       // council only, not base
	println("HasAny(dave):", g.HasAny(dave)) // somewhere in the group
	println("HasAny(zed):", g.HasAny(zed))   // nowhere

	// 3. Size (base only) vs TotalSize (deduplicated across everything).
	println("Size:", g.Size())
	println("TotalSize:", g.TotalSize())

	// 4. Iterate walks the base set; IterateAll walks base then roles in
	// name order, yielding each distinct address once.
	print("Iterate:")
	g.Iterate(0, 10, func(a address) bool {
		print(" " + string(a))
		return false
	})
	println()
	print("IterateAll:")
	g.IterateAll(0, 10, func(a address) bool {
		print(" " + string(a))
		return false
	})
	println()

	// 5. RolesContaining lists role names in lexicographic order; the
	// base set is not a role.
	println("RolesContaining(eve):", strings.Join(g.RolesContaining(eve), ","))
	println("RolesContaining(alice):", strings.Join(g.RolesContaining(alice), ","))

	// 6. Removing a role discards only the role: its members survive in
	// the base set and in other roles.
	g.RemoveRole("admin")
	println("after RemoveRole(admin):")
	println("Has(carol):", g.Has(carol))       // still a base member
	println("HasAny(eve):", g.HasAny(eve))     // still a moderator
	println("HasAny(grace):", g.HasAny(grace)) // was admin only, now gone
	println("TotalSize:", g.TotalSize())

	// 7. A readonly view mirrors every read and exposes no mutators; it
	// is the only handle meant to cross a realm boundary.
	rg := g.Readonly()
	println("readonly Size:", rg.Size())
	println("readonly HasAny(dave):", rg.HasAny(dave))
	rr, _ := rg.GetRole("moderator")
	println("readonly moderator members:", rr.Members().Size())
}

// Output:
// Has(alice): true
// Has(dave): false
// HasAny(dave): true
// HasAny(zed): false
// Size: 3
// TotalSize: 7
// Iterate: alice bob carol
// IterateAll: alice bob carol eve grace dave frank
// RolesContaining(eve): admin,moderator
// RolesContaining(alice):
// after RemoveRole(admin):
// Has(carol): true
// HasAny(eve): true
// HasAny(grace): false
// TotalSize: 6
// readonly Size: 3
// readonly HasAny(dave): true
// readonly moderator members: 2
