// PKGPATH: gno.land/r/tests/vm/recoverfvlibmade
package recoverfvlibmade

import "gno.land/p/demo/tests/recoverfv"

var n int

// A library closure made by this realm, run inside a borrowed library method,
// recovers a panic of the library's stored object: nothing on the way had
// write power over the library (master aborts), and the same route cannot write it.
func main(cur realm) {
	n = 1
	println("rec:", recoverfv.Shared().RunAny(recoverfv.Mk()))
	println("after", n)
	println("write:", recoverfv.Shared().RunAny(func() (rec any) {
		defer func() { rec = recover() }()
		recoverfv.StoredInc()
		return
	}))
	println("lib state:", recoverfv.Shared().N())
	println("OUTCOME caught")
}

// Output:
// rec: int is not of type string
// after 1
// write: cannot modify object owned by another realm from library code (call a crossing function of the owning realm): (const (ref(gno.land/p/demo/tests/recoverfv) package{})).shared.n
// lib state: 0
// OUTCOME caught
