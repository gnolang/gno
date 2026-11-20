// PKGPATH: gno.land/r/demo/entryescapedptr
package entryescapedptr

import "gno.land/r/tests/vm/viewdefault"

// An escaped local pointer to the owner's tree is still a view: a stranger
// writing through (*pt).Set writes nothing, and the owner's value stays put.
// The owner's own entry still writes.

func main(cur realm) {
	t := viewdefault.GetView()
	pt := &t
	func() {
		defer func() { println("escaped:", recover()) }()
		(*pt).Set("k", "pwned")
	}()
	println("view:", viewdefault.Read("k"))
	viewdefault.SetView(cross(cur), "k", "v1")
	println("owner wrote:", viewdefault.Read("k"))
}

// Output:
// escaped: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// view: v0
// owner wrote: v1
