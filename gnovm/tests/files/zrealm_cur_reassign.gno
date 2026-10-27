// PKGPATH: gno.land/r/demo/curreassign
package curreassign

// A crossing function's `cur` names the realm the frame is executing as. That
// binding is established at entry and fixed for the life of the frame, so
// rebinding the name makes it describe a realm the frame is not in. A helper
// entered without cross() takes its frame identity from the `cur` it is
// handed, so a rebound `cur` would carry that wrong answer into the callee,
// where IsCurrent() and cross() would both agree with it. Reject the rebind.

func probe(cur realm) {
	_ = cur.IsCurrent()
}

func main(cur realm) {
	cur = cur.Previous()
	probe(cur)
}

// Error:
// gno.land/r/demo/curreassign/zrealm_cur_reassign.gno:16:2-22: cannot reassign the crossing `cur` parameter: it names the realm this frame is executing as, and that binding is fixed for the life of the frame
