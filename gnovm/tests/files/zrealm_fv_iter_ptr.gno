// PKGPATH: gno.land/r/fviterptr
package fviterptr

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	n := 0
	fvx.PTree.Iterate("", "", func(k string, v any) bool {
		n++
		return false
	})
	println(n)
}
