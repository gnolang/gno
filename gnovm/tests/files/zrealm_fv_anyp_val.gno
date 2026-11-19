// PKGPATH: gno.land/r/fvanypval
package fvanypval

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	x := any(*fvx.Ptr())
	println(x != nil)
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer)

// TypeCheckError:
// gno.land/r/fvanypval/zrealm_fv_anyp_val.gno:7:11: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
