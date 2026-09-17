// PKGPATH: gno.land/r/demo/currange
package currange

// `for _, cur = range rs` writes the crossing parameter's slot without being
// an assignment statement and without taking an address, so neither the
// AssignStmt nor the RefExpr rejection sees it. It preprocessed, deployed, and
// reached the callee's frame Cur — caught only by the runtime backstop, at
// call time. Found by review AFTER the address-of rejection was added, which
// is why that backstop is kept rather than deleted as unreachable.

func steal(cur realm) {
	_ = cur.IsCurrent()
}

func main(cur realm) {
	rs := []realm{cur.Previous()}
	for _, cur = range rs {
	}
	steal(cur)
}

// Error:
// gno.land/r/demo/currange/zrealm_cur_range.gno:17:2-18:3: cannot assign to a realm-typed `cur` in a range clause: it is a caller-bound capability and rebinding it would forge the current-realm identity
