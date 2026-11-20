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

// Error:
// cannot allocate gno.land/r/tests/vm/launderrarrnested.N in realm gno.land/r/launder_rdata_arrdefault_nested_owner

// TypeCheckError:
// gno.land/r/launder_rdata_arrdefault_nested_owner/zrealm_launder_rdata_arrdefault_nested_owner.gno:8:6: cannot hold a value of gno.land/r/tests/vm/launderrarrnested.N outside its realm (use a pointer); gno.land/r/launder_rdata_arrdefault_nested_owner/zrealm_launder_rdata_arrdefault_nested_owner.gno:9:2: cannot hold a value of gno.land/r/tests/vm/launderrarrnested.N outside its realm (use a pointer); gno.land/r/launder_rdata_arrdefault_nested_owner/zrealm_launder_rdata_arrdefault_nested_owner.gno:10:2: cannot hold a value of gno.land/r/tests/vm/launderrarrnested.S outside its realm (use a pointer); gno.land/r/launder_rdata_arrdefault_nested_owner/zrealm_launder_rdata_arrdefault_nested_owner.gno:11:2: cannot hold a value of gno.land/r/tests/vm/launderrarrnested.P outside its realm (use a pointer)
