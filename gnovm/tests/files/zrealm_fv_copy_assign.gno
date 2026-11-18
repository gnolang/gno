// PKGPATH: gno.land/r/fvcopyassign
package fvcopyassign

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	p := fvx.Ptr()
	c := *p
	println(c.N)
}
