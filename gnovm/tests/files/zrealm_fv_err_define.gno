// PKGPATH: gno.land/r/fverrdefine
package fverrdefine

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	err := fvx.Fail()
	println(err != nil)
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.E outside realm gno.land/r/tests/vm/fvx (use a pointer)
