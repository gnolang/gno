// PKGPATH: gno.land/r/fvarranon
package fvarranon

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	p := new([1]fvx.T)
	println(p[0].N)
}

// Error:
// cannot allocate gno.land/r/tests/vm/fvx.T in realm gno.land/r/fvarranon
