// PKGPATH: gno.land/r/fvrangeblank
package fvrangeblank

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	n := 0
	for i, _ := range fvx.Arr {
		n += i
	}
	println(n)
}

// Output:
// 1
