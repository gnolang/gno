// PKGPATH: gno.land/r/fvrecover
package fvrecover

import "gno.land/r/tests/vm/fvx"

var saved any

func main(cur realm) {
	defer func() {
		saved = recover()
		println(saved != nil)
	}()
	fvx.Panics()
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer)
