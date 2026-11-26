// PKGPATH: gno.land/r/tests/vm/recoverfvpborrowwrite
package recoverfvpborrowwrite

import "gno.land/p/demo/tests/recoverfv"

// Twin of recover_fv_pborrow: the same route writing the library's object is
// refused after init, and the refusal cannot be recovered.
func main(cur realm) {
	c := recoverfv.Shared()
	ops := []func(){func() { c.Inc() }}
	rec := func() (rec any) {
		defer func() { rec = recover() }()
		ops[0]()
		return
	}()
	println("not reached", rec)
}

// Error:
// cannot mutate gno.land/p/demo/tests/recoverfv: package is immutable post-init
