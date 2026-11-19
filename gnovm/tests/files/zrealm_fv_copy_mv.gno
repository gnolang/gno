// PKGPATH: gno.land/r/fvcopymv
package fvcopymv

import "gno.land/r/tests/vm/fvx"

var kf func() int

func main(cur realm) {
	kf = fvx.Ptr().Val
	println(kf())
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer)
