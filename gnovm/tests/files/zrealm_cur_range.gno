// PKGPATH: gno.land/r/demo/currange
package currange

// `for _, cur = range rs` writes the crossing parameter's slot without being
// an assignment statement and without taking an address, so neither the
// AssignStmt nor the RefExpr rejection sees it. It preprocessed and reached the
// callee's frame identity, leaving only the runtime check to notice, at call
// time. Found by review AFTER the address-of rejection was added — which is
// the argument for keeping that runtime check rather than deleting it as
// unreachable: this is the third write shape into one slot, and the second one
// the syntactic rules missed.

func probe(cur realm) {
	_ = cur.IsCurrent()
}

func main(cur realm) {
	rs := []realm{cur.Previous()}
	for _, cur = range rs {
	}
	probe(cur)
}

// Error:
// gno.land/r/demo/currange/zrealm_cur_range.gno:19:2-20:3: cannot assign to a realm-typed `cur` in a range clause: it names the realm this frame is executing as, and that binding is fixed for the life of the frame
