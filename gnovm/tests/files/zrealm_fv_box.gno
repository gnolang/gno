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

// Output:
// 2
