// PKGPATH: gno.land/r/tests/vm/recoverfvkept
package recoverfvkept

import (
	"gno.land/p/demo/tests/recoverfv"
	"gno.land/r/tests/vm/recoverfvb"
)

// A library-stored closure kept by realm B and run by B recovers, as on
// master.

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

func main(cur realm) {
	recoverfvb.Keep(cross(cur), recoverfv.StoredTry)
	println("rec:", recoverfvb.RunKept())
	println("OUTCOME caught")
}

// Output:
// rec: int is not of type string
// OUTCOME caught
