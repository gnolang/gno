// PKGPATH: gno.land/r/fvanyp
package fvanyp

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	x := any(fvx.Ptr())
	println(x.(*fvx.T).N)
}

// Output:
// 1
