// PKGPATH: gno.land/r/test/deferfix1
package deferfix1

// Test 2: Same-realm defer-method (regression sanity).
// recvOID.PkgID == m.Realm.ID, so the implicit borrow-switch is a
// no-op. Pre-fix and post-fix both pass; this guards against the
// fix breaking the no-switch path.

type Counter struct {
	N int
}

func (c *Counter) Inc() {
	c.N++
}

// Persisted in this realm.
var c = &Counter{N: 0}

func main(cur realm) {
	defer c.Inc()
	println("c.N before defer fires =", c.N)
}

// Output:
// c.N before defer fires = 0
