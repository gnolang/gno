// PKGPATH: gno.land/r/fvmiss
package fvmiss

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	var x any = 1
	_, ok := x.(fvx.T)
	println(fvx.M["nope"].N, ok)
}

// Output:
// 0 false

// TypeCheckError:
// gno.land/r/fvmiss/zrealm_fv_miss.gno:8:11: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
