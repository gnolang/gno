// PKGPATH: gno.land/r/test/methodexprinline
package methodexprinline

// Probe: `Type.Method(recv)` (method expression) on a cross-realm-stored
// receiver, where the receiver was constructed in the caller's realm and
// then handed off to crossrealm for storage. Under interrealm v2 Phase 2
// (storage = authority, stamped at construction), recv.PkgID stays at
// the constructor's realm even after being saved elsewhere, so
// IsReadonlyBy() does not panic when the caller mutates a field directly.
// The println therefore runs and the test succeeds.

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
	crossrealm.SetObject(cross2(cur), s)
	(*S).Close(s) // inline method-expression call
	println("should not print")
}

// Output:
// should not print
