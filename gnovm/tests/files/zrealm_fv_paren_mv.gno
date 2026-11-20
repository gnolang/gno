// PKGPATH: gno.land/r/fvparenmv
package fvparenmv

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	p := fvx.Ptr()
	println((p.Val)())
}

// Output:
// 1
