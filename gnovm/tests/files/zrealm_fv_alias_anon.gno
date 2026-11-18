// PKGPATH: gno.land/r/fvaliasanon
package fvaliasanon

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	var s fvx.S
	println(s.N)
}

// Error:
// cannot allocate struct{N int} (declared in gno.land/r/tests/vm/fvx) in realm gno.land/r/fvaliasanon
