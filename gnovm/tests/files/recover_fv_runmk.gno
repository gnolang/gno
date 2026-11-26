// PKGPATH: gno.land/r/tests/vm/recoverfvrunmk
package recoverfvrunmk

import "gno.land/p/demo/tests/recoverfv"

// A library closure made by this realm runs inside a library-stored closure
// and recovers the library object's panic (master aborts).

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

func main(cur realm) {
	mk := recoverfv.Mk()
	ops := []func(func() any) any{recoverfv.StoredRun}
	runInc := ops
	println("rec:", ops[0](mk))
	println("after")
	println("write:", try(func() { runInc[0](func() any { recoverfv.StoredInc(); return nil }) }))
	println("lib state:", recoverfv.Shared().N())
	println("OUTCOME caught")
}

// Output:
// rec: int is not of type string
// after
// write: cannot modify object owned by another realm from library code (call a crossing function of the owning realm): (const (ref(gno.land/p/demo/tests/recoverfv) package{})).shared.n
// lib state: 0
// OUTCOME caught
