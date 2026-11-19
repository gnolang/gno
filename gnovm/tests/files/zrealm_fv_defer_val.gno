// PKGPATH: gno.land/r/fvdeferval
package fvdeferval

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	p := fvx.Ptr()
	defer p.Val()
	println("ok")
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer)
