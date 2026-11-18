// PKGPATH: gno.land/r/fvitervalue
package fvitervalue

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	n := 0
	fvx.Tree.Iterate("", "", func(k string, v any) bool {
		n++
		return false
	})
	println(n)
}
