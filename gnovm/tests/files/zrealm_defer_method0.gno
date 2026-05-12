// PKGPATH: gno.land/r/test/deferfix0
package deferfix0

// Test 1: Core fix. defer-method on a struct that's been persisted in
// another realm must perform the implicit borrow-realm switch (just like
// inline method calls do). Pre-fix this panicked with "cannot directly
// modify readonly tainted object". Post-fix, the mutation succeeds.

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

	defer c.Inc()
	println("before defer fires, c.N =", c.N)
}

// Output:
// before defer fires, c.N = 0

// Realm:
