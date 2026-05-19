// PKGPATH: gno.land/r/test/deferfixrecover
package deferfixrecover

// Test 7: defer-method on a constructor-owned struct, with an outer
// defer-recover that would have swallowed any readonly-taint panic
// under the old finalize-time-ownership model. Under interrealm v2
// Phase 2 the struct keeps its constructor PkgID even after being
// stored in another realm, so the deferred `s.Close()` mutates s
// directly — no panic raised, recover sees nothing to swallow.

import (
	crossrealm "gno.land/r/tests/vm/crossrealm"
)

type Switch struct {
	Closed bool
}

func (s *Switch) Close() {
	s.Closed = true
}

func main(cur realm) {
	defer func() {
		if r := recover(); r != nil {
			println("RECOVERED:", r)
		}
	}()

	s := &Switch{}
	crossrealm.SetObject(cross(cur), s)

	defer s.Close()

	println("Closed before defer fires:", s.Closed)
}

// Output:
// Closed before defer fires: false
