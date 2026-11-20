// PKGPATH: gno.land/r/fvcall
package fvcall

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	println(fvx.Ptr().Val(), fvx.Get().N)
}

// Output:
// 1 1
