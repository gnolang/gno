// PKGPATH: gno.land/r/fvcopymv
package fvcopymv

import "gno.land/r/tests/vm/fvx"

var kf func() int

func main(cur realm) {
	kf = fvx.Ptr().Val
	println(kf())
}
