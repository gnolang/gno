// PKGPATH: gno.land/r/demo/curaddr
package curaddr

// The address of a crossing function's `cur` may not be taken. Rejecting
// `cur = cur.Previous()` matches only a bare name on the left of an
// assignment; `p := &cur; *p = cur.Previous()` writes the same parameter
// slot through a StarExpr, matches no such rule, and used to preprocess and
// deploy cleanly — reaching the callee's frame Cur and being stopped only by
// doOpCall's runtime backstop, at call time rather than at deploy time.
// A slot whose address cannot be taken cannot be written through any
// assignment shape, so the address is what is refused.

func steal(cur realm) {
	_ = cur.IsCurrent()
}

func main(cur realm) {
	p := &cur
	*p = cur.Previous()
	steal(cur)
}

// Error:
// gno.land/r/demo/curaddr/zrealm_cur_addr.gno:18:7-11: cannot take the address of a realm-typed `cur`: a pointer to it could be used to rebind it and forge the current-realm identity. `realm` is an interface, so a *realm is never needed — pass `cur` by value
