// PKGPATH: gno.land/r/fvmethexpr
package fvmethexpr

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	p := fvx.Ptr()
	println(fvx.T.Val(*p))
}

// Output:
// 1

// TypeCheckError:
// gno.land/r/fvmethexpr/zrealm_fv_methexpr.gno:8:20: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
