// PKGPATH: gno.land/r/fvanonargnc
package fvanonargnc

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	println(fvx.OptsNC(struct{ N int }{N: 5}))
}

// Output:
// 5
