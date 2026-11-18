// PKGPATH: gno.land/r/fvcopyappend
package fvcopyappend

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	s := make([]fvx.T, 0)
	s = append(s, fvx.M["nope"])
	p := &s[0]
	p.N = 5
	println(fvx.Read(cross(cur), p))
}
