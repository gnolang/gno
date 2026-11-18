// PKGPATH: gno.land/r/fvcopyany
package fvcopyany

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	var x any = *fvx.Ptr()
	println(x != nil)
}
