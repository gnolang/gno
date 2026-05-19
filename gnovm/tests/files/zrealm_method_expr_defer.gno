// PKGPATH: gno.land/r/test/methodexprdefer
package methodexprdefer

// Probe: `defer Type.Method(recv)` on a cross-realm-stored receiver
// constructed in the caller's realm. Under interrealm v2 Phase 2
// (storage = authority, stamped at construction), recv.PkgID stays at
// the constructor's realm even after the SetObject hand-off, so the
// deferred mutation runs without a readonly-taint panic. (Note: this
// previously also probed the borrow-realm-switch in deferred-method
// dispatch — that probe no longer triggers under the construction-
// authority model because the constructor's realm matches recv.PkgID.)

import (
	crossrealm "gno.land/r/tests/vm/crossrealm"
)

type S struct {
	Closed bool
}

func (s *S) Close() {
	s.Closed = true
}

func main(cur realm) {
	s := &S{}
	crossrealm.SetObject(cross(cur), s)
	defer (*S).Close(s)
	println("body done")
}

// Output:
// body done
