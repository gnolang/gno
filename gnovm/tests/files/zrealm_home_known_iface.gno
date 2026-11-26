// PKGPATH: gno.land/r/zrealm_home_known_iface
package zrealm_home_known_iface

// R2a: the gate reaches cbvictim's own value through an interface this
// realm supplied; the value was made in another crossing call, so it is read-only here.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunAny(cross(cur), cbvictim.Obj)
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
