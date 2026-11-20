// PKGPATH: gno.land/r/fvanypnil
package fvanypnil

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	var x fvx.I = fvx.I(nil)
	println(x == nil)
}

// Output:
// true
