// PKGPATH: gno.land/r/fvbox
package fvbox

import "gno.land/r/tests/vm/fvx"

var saved []any

func main(cur realm) {
	var i any
	for _, i = range fvx.Arr {
		saved = append(saved, i)
	}
	println(len(saved))
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer)
