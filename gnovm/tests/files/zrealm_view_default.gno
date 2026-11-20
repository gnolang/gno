// PKGPATH: gno.land/r/demo/viewuser
package viewuser

import "gno.land/r/tests/vm/viewdefault"

// A returned reference is a view: reads work, a /p/ method write is refused.
// The stranger never entered the owner's code, so the write hits the readonly
// gate.

func main(cur realm) {
	t := viewdefault.GetView()
	println("read:", t.Get("k"))
	t.Set("k", "pwned")
}

// Output:
// read: v0

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
