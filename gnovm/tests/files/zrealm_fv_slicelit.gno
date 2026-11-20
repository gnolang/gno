// PKGPATH: gno.land/r/fvslicelit
package fvslicelit

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	p := &[]fvx.T{1: fvx.Get()}[0]
	p.N = 5
	println(fvx.Read(cross(cur), p))
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer)

// TypeCheckError:
// gno.land/r/fvslicelit/zrealm_fv_slicelit.gno:7:19: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
