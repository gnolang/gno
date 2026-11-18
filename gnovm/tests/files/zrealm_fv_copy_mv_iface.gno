// PKGPATH: gno.land/r/fvcopymviface
package fvcopymviface

import "gno.land/r/tests/vm/fvx"

var kf func() string

func main(cur realm) {
	kf = fvx.Fail().Error
	println(kf())
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.E outside realm gno.land/r/tests/vm/fvx (use a pointer)
