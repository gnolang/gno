// PKGPATH: gno.land/r/zrealm_home_iface_arg
package zrealm_home_iface_arg

// A function value inside an interface argument is judged like a bare one.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunIface(cross(cur), any(cbvictim.FV))
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>
