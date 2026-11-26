// PKGPATH: gno.land/r/tests/vm/recoverfvlibnest
package recoverfvlibnest

import "gno.land/p/demo/tests/recoverfv"

// Same as recover_fv_libmade with two library closures between the recover
// and the panic.
func main(cur realm) {
	println("rec:", recoverfv.Shared().RunAny(recoverfv.Mk2()))
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
// write: cannot modify object owned by another realm from library code (call a function of the owning realm): (const (ref(gno.land/p/demo/tests/recoverfv) package{})).shared.n
// lib state: 0
// OUTCOME caught
