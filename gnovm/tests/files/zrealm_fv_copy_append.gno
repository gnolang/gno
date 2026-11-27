// PKGPATH: gno.land/r/fvcopyappend
package fvcopyappend

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	s := make([]fvx.T, 0)
	s = append(s, fvx.M["nope"])
	p := &s[0]
	p.N = 5
	println(fvx.Read(cross(cur), p))
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer)

// TypeCheckError:
// gno.land/r/fvcopyappend/zrealm_fv_copy_append.gno:7:2: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvcopyappend/zrealm_fv_copy_append.gno:7:7: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvcopyappend/zrealm_fv_copy_append.gno:8:2: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvcopyappend/zrealm_fv_copy_append.gno:8:6: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvcopyappend/zrealm_fv_copy_append.gno:8:13: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvcopyappend/zrealm_fv_copy_append.gno:8:16: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
