// PKGPATH: gno.land/r/demo/curaddr
package curaddr

// The address of a crossing function's `cur` may not be taken. Rejecting
// `cur = ...` matches only a bare name on the left of an assignment;
// `p := &cur; *p = cur.Previous()` writes the same parameter slot through a
// StarExpr and matches no such rule, so it preprocessed and reached the
// callee's frame identity, leaving only doOpCall's runtime check to notice —
// at call time rather than at compile time. A slot whose address cannot be
// taken cannot be written through any assignment shape, so the address is what
// is refused.

func probe(cur realm) {
	_ = cur.IsCurrent()
}

func main(cur realm) {
	p := &cur
	*p = cur.Previous()
	probe(cur)
}

// Error:
// gno.land/r/demo/curaddr/zrealm_cur_addr.gno:18:7-11: cannot take the address of a realm-typed `cur`: the binding is fixed for the life of the frame, and a pointer to it is a way to rebind it. `realm` is an interface, so a *realm is never needed — pass `cur` by value
