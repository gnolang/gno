// PKGPATH: gno.land/r/fvslicecap
package fvslicecap

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	s := make([]fvx.T, 0, 2)
	println(cap(s))
}

// Error:
// cannot allocate gno.land/r/tests/vm/fvx.T in realm gno.land/r/fvslicecap

// TypeCheckError:
// gno.land/r/fvslicecap/zrealm_fv_slice_cap.gno:7:2: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvslicecap/zrealm_fv_slice_cap.gno:7:7: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
