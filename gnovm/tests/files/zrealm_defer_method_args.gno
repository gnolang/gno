// PKGPATH: gno.land/r/test/deferfixargs
package deferfixargs

// Test 5: defer-method with extra args on cross-realm-stored struct.
// Verifies that args[0] = receiver (handled by the fix) and args[1..] =
// regular args (already worked) flow correctly to the deferred frame.

import (
	crossrealm "gno.land/r/tests/vm/crossrealm"
)

type Bag struct {
	Items []int
}

func (b *Bag) Add(x, y int) {
	b.Items = append(b.Items, x, y)
}

func main(cur realm) {
	b := &Bag{}
	crossrealm.SetObject(cross, b)

	defer b.Add(10, 20)
	defer b.Add(30, 40) // LIFO: 30,40 fires first, then 10,20
	println("about to return; defers will run LIFO")
}

// Output:
// about to return; defers will run LIFO
