// PKGPATH: gno.land/r/demo/curaddrarg
package curaddrarg

// The helper form of the address-of route: `&cur` never appears on the left of
// anything, it is handed to a function that writes through the pointer. No LHS
// rule could see this, which is the argument for refusing the ADDRESS rather
// than enumerating assignment shapes. Kept separate from zrealm_cur_addr.gno so
// that narrowing the RefExpr rule to assignment context fails a test instead of
// silently reopening the route.

func set(p *realm, v realm) {
	*p = v
}

func probe(cur realm) {
	_ = cur.IsCurrent()
}

func main(cur realm) {
	set(&cur, cur.Previous())
	probe(cur)
}

// Error:
// gno.land/r/demo/curaddrarg/zrealm_cur_addr_arg.gno:20:6-10: cannot take the address of a realm-typed `cur`: the binding is fixed for the life of the frame, and a pointer to it is a way to rebind it. `realm` is an interface, so a *realm is never needed — pass `cur` by value
