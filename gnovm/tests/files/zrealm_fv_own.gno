// PKGPATH: gno.land/r/fvown
package fvown

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	p := fvx.New(3)
	println(p.N, fvx.Ptr().N, fvx.Arr[1].N)
}

// Output:
// 3 1 200
