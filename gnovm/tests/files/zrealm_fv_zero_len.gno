// PKGPATH: gno.land/r/fvzerolen
package fvzerolen

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	var x [0]fvx.T
	println(len(x))
}

// Output:
// 0

// TypeCheckError:
// gno.land/r/fvzerolen/zrealm_fv_zero_len.gno:7:6: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
