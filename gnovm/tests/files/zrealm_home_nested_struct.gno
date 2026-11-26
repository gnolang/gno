// PKGPATH: gno.land/r/zrealm_home_nested_struct
package zrealm_home_nested_struct

// A struct copied into the gate's argument copies its fields under this
// realm's hold: the function field arrives without a maker.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunNested(cross(cur), cbvictim.Nested)
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
