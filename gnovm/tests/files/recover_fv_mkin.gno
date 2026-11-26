// PKGPATH: gno.land/r/tests/vm/recoverfvmkin
package recoverfvmkin

import "gno.land/p/demo/tests/recoverfv"

// A closure made inside a borrowed library method, called by value,
// recovers the library object's panic, as on master.

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

func main(cur realm) {
	f := recoverfv.Shared().MkIn()
	println("rec:", f())
	println("OUTCOME caught")
}

// Output:
// rec: int is not of type string
// OUTCOME caught
