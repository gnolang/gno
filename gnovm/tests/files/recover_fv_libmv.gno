// PKGPATH: gno.land/r/tests/vm/recoverfvlibmv
package recoverfvlibmv

import "gno.land/p/demo/tests/recoverfv"

// Same as recover_fv_libmade with a method value bound in this realm.
func main(cur realm) {
	h := &recoverfv.H{}
	println("rec:", recoverfv.Shared().RunAny(h.Try))
	println("write:", try(func() { recoverfv.Shared().RunAny(recoverfv.Shared().Inc) }))
	println("lib state:", recoverfv.Shared().N())
	println("OUTCOME caught")
}

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

// Output:
// rec: int is not of type string
// write: cannot modify object owned by another realm from library code (call a crossing function of the owning realm): c<VPBlock(1,0)>.n
// lib state: 0
// OUTCOME caught
