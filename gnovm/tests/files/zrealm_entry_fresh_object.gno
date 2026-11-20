// PKGPATH: gno.land/r/demo/entryfresh
package entryfresh

import "gno.land/r/tests/vm/entries"

// An object allocated in this transaction is still the owner's object: a
// foreign caller cannot write it. The owner's own entry writes it.

func main(cur realm) {
	t := entries.NewTree()
	func() {
		defer func() { println("fresh foreign:", recover()) }()
		t.Set("k", "pwned")
	}()
	println("fresh empty:", entries.ReadFrom(t, "k") == "")
	entries.OwnerSetTree(cross(cur), t, "k", "v1")
	println("owner wrote:", entries.ReadFrom(t, "k"))
}

// Output:
// fresh foreign: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// fresh empty: true
// owner wrote: v1
