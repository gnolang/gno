// PKGPATH: gno.land/r/demo/entryphandle
package entryphandle

import "gno.land/r/tests/vm/viewdefault"

// A /p/-built writer onto the owner's view is library code: a stranger calling
// Set writes nothing. The owner's own entry still writes its view.

func main(cur realm) {
	func() {
		defer func() { println("writer:", recover()) }()
		viewdefault.NewViewWriter().Set("k", "via-writer")
	}()
	println("view:", viewdefault.Read("k"))
	viewdefault.SetView(cross(cur), "k", "v1")
	println("owner wrote:", viewdefault.Read("k"))
}

// Output:
// writer: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// view: v0
// owner wrote: v1
