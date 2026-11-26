// PKGPATH: gno.land/r/zrealm_home_panic
package zrealm_home_panic

// A panic value is copied by this realm's code: cbvictim's value it carries
// arrives without a maker, so the gate's recover runs it read-only.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunPanic(cross(cur), func() { panic(cbvictim.FV) })
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
