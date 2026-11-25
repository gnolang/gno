// PKGPATH: gno.land/r/launder_rdata_arrdefault_nested_owner
package launder_rdata_arrdefault_nested_owner

// The declaring realm can still write to the nested arrays it was handed.

import v "gno.land/r/tests/vm/launderrarrnested"

type Local struct {
	N v.N
	S v.S
	P v.P
}

var g = &Local{}

func main(cur realm) {
	v.RegisterN(&g.N)
	v.RegisterS(&g.S)
	v.RegisterP(&g.P)
	v.SetS(cross(cur), 6)
	println("after owner:", v.Trusted())
}

// Output:
// after owner: 0 6 0
