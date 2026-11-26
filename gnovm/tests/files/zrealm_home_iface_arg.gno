// PKGPATH: gno.land/r/zrealm_home_iface_arg
package zrealm_home_iface_arg

// A function value inside an interface argument is judged like a bare one.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunIface(cross(cur), any(cbvictim.FV))
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
