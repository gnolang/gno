// PKGPATH: gno.land/r/demo/viewuser
package viewuser

import "gno.land/r/tests/vm/viewdefault"

// A returned reference is a view: reads work, a /p/ method write is refused.
// The owner never wrote mutable(view), so borrow rule #2 does not switch
// storage to the owner and the write hits the readonly gate.

func main(cur realm) {
	t := viewdefault.GetView()
	println("read:", t.Get("k"))
	t.Set("k", "pwned")
}

// Output:
// read: v0

// Error:
// cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): tree<VPBlock(1,0)>.node
