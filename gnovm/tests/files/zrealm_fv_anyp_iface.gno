// PKGPATH: gno.land/r/fvanypiface
package fvanypiface

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	println(fvx.I(fvx.Ptr()).Val())
}

// Output:
// 1
