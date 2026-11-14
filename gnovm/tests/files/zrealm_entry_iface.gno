// PKGPATH: gno.land/r/demo/entryiface
package entryiface

import "gno.land/r/tests/vm/entries"

// A stored tree reached through an interface is still the owner's object: the
// foreign interface call writes nothing. The owner's own entry writes it.

func main(cur realm) {
	entries.OwnerSetIface(cross(cur), "k", "v1")
	println("owner wrote:", entries.ReadIface("k"))
	func() {
		defer func() { println("iface:", recover()) }()
		entries.AsSetter().Set("k", "pwned")
	}()
	println("after foreign:", entries.ReadIface("k"))
}

// Output:
// owner wrote: v1
// iface: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// after foreign: v1
