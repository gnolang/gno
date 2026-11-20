// PKGPATH: gno.land/r/demo/entrypromoted
package entrypromoted

import "gno.land/r/tests/vm/entries"

// A /p/ method promoted through an embedded value is library code: a foreign
// caller cannot write the owner's tree through it. The owner's own entry does.

func main(cur realm) {
	entries.OwnerSetHolder(cross(cur), "k", "v1")
	println("owner wrote:", entries.ReadHolder("k"))
	func() {
		defer func() { println("promoted:", recover()) }()
		entries.GetHolder().Set("k", "pwned")
	}()
	println("after foreign:", entries.ReadHolder("k"))
}

// Output:
// owner wrote: v1
// promoted: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// after foreign: v1
