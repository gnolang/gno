// PKGPATH: gno.land/r/zrealm_home_recv
package zrealm_home_recv

// A function-typed receiver is handed over like an argument: cbvictim.H
// (maker cbvictim) runs read-only in its own crossing method.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.H.Run(cross(cur))
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
