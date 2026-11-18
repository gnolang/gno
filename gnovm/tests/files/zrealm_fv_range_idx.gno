// PKGPATH: gno.land/r/fvrangeidx
package fvrangeidx

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	n := 0
	for i := range fvx.Arr {
		n += fvx.Arr[i].N
	}
	for i := range fvx.Sl {
		n += fvx.Sl[i].N
	}
	println(n)
}

// Output:
// 303
