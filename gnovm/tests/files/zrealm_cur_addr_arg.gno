// PKGPATH: gno.land/r/demo/curaddrarg
package curaddrarg

// The helper form of the address-of route: `&cur` never appears on the left of
// anything, it is handed to a function that writes through the pointer. No LHS
// rule could see this, which is the argument for refusing the ADDRESS rather
// than chasing assignment shapes. Kept separate from zrealm_cur_addr.gno so
// that narrowing the RefExpr rule to assignment context fails a test instead of
// silently reopening the route.

func set(p *realm, v realm) {
	*p = v
}

func steal(cur realm) {
	_ = cur.IsCurrent()
}

func main(cur realm) {
	set(&cur, cur.Previous())
	steal(cur)
}

// Error:
// gno.land/r/demo/curaddrarg/zrealm_cur_addr_arg.gno:20:6-10: cannot take the address of a realm-typed `cur`: a pointer to it could be used to rebind it and forge the current-realm identity. `realm` is an interface, so a *realm is never needed — pass `cur` by value
