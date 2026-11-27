// PKGPATH: gno.land/r/zrealm_handle_mexpr
package zrealm_handle_mexpr

// A method expression taken as a value in the victim's read-only code writes
// nothing in its gate; one taken in a gate and handed back writes nothing too.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	println("read-only:", revive(func() { cbvictim.RunM(cross(cur), cbvictim.GetMExpr()) }))
	println("name: [" + cbvictim.Box.Name + "]")
	println("gate:", revive(func() { cbvictim.RunM(cross(cur), cbvictim.GateMExpr(cross(cur))) }))
	println("name: [" + cbvictim.Box.Name + "]")
}

// Output:
// read-only: gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): t<VPBlock(1,0)>.Name
// name: []
// gate: gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): t<VPBlock(1,0)>.Name
// name: []
