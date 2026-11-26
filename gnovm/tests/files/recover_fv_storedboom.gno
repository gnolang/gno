// PKGPATH: gno.land/r/tests/vm/recoverfvstoredboom
package recoverfvstoredboom

import "gno.land/p/demo/tests/recoverfv"

// A library-stored closure called by value panics on the library's object;
// it had no write power over the library, so this realm recovers (master
// aborts). The same route cannot write the library.

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

var n int

func main(cur realm) {
	n = 1
	ops := []func(func()){func(f func()) { f() }}
	println("rec:", try(func() { ops[0](recoverfv.StoredBoom) }))
	println("after", n)
	println("write:", try(func() { ops[0](recoverfv.StoredInc) }))
	println("lib state:", recoverfv.Shared().N())
	println("OUTCOME caught")
}

// Output:
// rec: int is not of type string
// after 1
// write: cannot modify object owned by another realm from library code (call a function of the owning realm): (const (ref(gno.land/p/demo/tests/recoverfv) package{})).shared.n
// lib state: 0
// OUTCOME caught
