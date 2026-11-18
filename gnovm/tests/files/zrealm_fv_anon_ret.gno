// PKGPATH: gno.land/r/fvanonret
package fvanonret

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	s := fvx.Stats()
	println(s.A + s.B)
	a := fvx.Pair()
	println(a[0])
}

// Output:
// 3
// 3
