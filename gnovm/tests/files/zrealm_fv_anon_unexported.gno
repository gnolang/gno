// PKGPATH: gno.land/r/fvanonunexported
package fvanonunexported

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	var x any = fvx.Hidden()
	println(x != nil)
}

// Error:
// cannot copy struct{n int} outside realm gno.land/r/tests/vm/fvx (use a pointer)
