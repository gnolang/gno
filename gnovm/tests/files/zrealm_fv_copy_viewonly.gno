// PKGPATH: gno.land/r/fvcopyviewonly
package fvcopyviewonly

import (
	"gno.land/p/demo/tests/pboxget"
	"gno.land/r/tests/vm/fvx"
)

var saved any

func main(cur realm) {
	f := pboxget.First // a function value: its call runs view-only
	saved = f(fvx.Boxes)
	println(saved != nil)
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer)
