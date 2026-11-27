// PKGPATH: gno.land/r/fvcopyassign
package fvcopyassign

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	p := fvx.Ptr()
	c := *p
	println(c.N)
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer)

// TypeCheckError:
// gno.land/r/fvcopyassign/zrealm_fv_copy_assign.gno:8:2: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvcopyassign/zrealm_fv_copy_assign.gno:8:7: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
