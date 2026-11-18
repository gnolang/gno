// PKGPATH: gno.land/r/fvcopyany
package fvcopyany

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	var x any = *fvx.Ptr()
	println(x != nil)
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer)

// TypeCheckError:
// gno.land/r/fvcopyany/zrealm_fv_copy_any.gno:7:14: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
