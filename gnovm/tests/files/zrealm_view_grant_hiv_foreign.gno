// PKGPATH: gno.land/r/demo/hivforeign
package hivforeign

import "gno.land/r/tests/vm/viewdefault"

// t is an escaped local of pointer type, so &t is a pointer whose base is a
// heap item owned by THIS realm whose wrapped value is a pointer into
// viewdefault's storage. GetGrantObject descends that heap item and reaches
// the foreign tree; mutable must still refuse, and the grant must not exist.
func main(cur realm) {
	t := viewdefault.GetView()
	pt := &t
	func() {
		defer func() { println("mutable:", recover()) }()
		mutable(pt)
		println("mutable: ok")
	}()
	func() {
		defer func() { println("write:", recover()) }()
		t.Set("k", "pwned")
		println("write: ok")
	}()
}

// Output:
// mutable: mutable: only the owning realm may grant write access to its object
// write: cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): tree<VPBlock(1,0)>.node
