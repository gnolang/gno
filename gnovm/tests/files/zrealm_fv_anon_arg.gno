// PKGPATH: gno.land/r/fvanonarg
package fvanonarg

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	println(fvx.Opts(cross(cur), struct{ N int }{N: 5}))
}

// Output:
// 5
