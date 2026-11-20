// PKGPATH: gno.land/r/demo/entrydefer
package entrydefer

import "gno.land/r/tests/vm/entries"

// A deferred write runs with the authority of the code that deferred it: the
// owner's defer writes, a foreign caller's defer is refused.

func main(cur realm) {
	entries.OwnerSetDeferred(cross(cur), "k", "v1")
	println("owner deferred:", entries.Read("k"))
	func() {
		defer func() { println("foreign deferred:", recover()) }()
		defer entries.GetTree().Set("k", "pwned")
	}()
	println("after foreign:", entries.Read("k"))
}

// Output:
// owner deferred: v1
// foreign deferred: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// after foreign: v1
