// PKGPATH: gno.land/r/test/deferfixmultihop
package deferfixmultihop

// Test 10: multi-hop cross-realm. struct created here, persisted in
// crossrealm via cross-call, then defer-method runs from this realm's
// frame after the call returns. The borrow-switch must correctly
// target crossrealm's PkgID (where the struct now lives), not this
// realm's PkgID.

import (
	crossrealm "gno.land/r/tests/vm/crossrealm"
)

type Counter struct {
	N int
}

func (c *Counter) Inc() {
	c.N++
}

func main(cur realm) {
	c := &Counter{N: 0}
	crossrealm.SetObject(cross, c)
	// Re-fetch via getter to confirm crossrealm has it.
	stored := crossrealm.GetObject().(*Counter)
	defer stored.Inc()
	println("stored.N before defer fires =", stored.N)
}

// Output:
// stored.N before defer fires = 0
