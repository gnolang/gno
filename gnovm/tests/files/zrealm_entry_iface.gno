// PKGPATH: gno.land/r/demo/entryiface
package entryiface

import "gno.land/r/tests/vm/entries"

// A stored tree reached through an interface is still the owner's object: the
// foreign interface call runs view-only and aborts. The owner's entry writes it.

func main(cur realm) {
	entries.OwnerSetIface(cross(cur), "k", "v1")
	println("owner wrote:", entries.ReadIface("k"))
	entries.AsSetter().Set("k", "pwned")
}

// Output:
// owner wrote: v1

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
