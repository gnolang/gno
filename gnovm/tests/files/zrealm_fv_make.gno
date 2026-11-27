// PKGPATH: gno.land/r/fvmake
package fvmake

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	p := &make([]fvx.T, 1)[0]
	p.N = 5
	println(fvx.Read(cross(cur), p))
}

// Error:
// cannot allocate gno.land/r/tests/vm/fvx.T in realm gno.land/r/fvmake
