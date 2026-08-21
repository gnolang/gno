// PKGPATH: gno.land/r/test/deferfix0
package deferfix0

// Test 1: defer-method on a struct constructed in the caller's realm
// and then handed to another realm for storage. Storage is authority
// and is stamped at construction, so the struct's PkgID stays at the
// constructor's realm. Direct field mutation from that same realm
// succeeds — no readonly-tainted panic, regardless of where the
// object is currently persisted.

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
	crossrealm.SetObject(cross(cur), c)

	defer c.Inc()
	println("before defer fires, c.N =", c.N)
}

// Output:
// before defer fires, c.N = 0
