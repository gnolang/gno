// PKGPATH: gno.land/r/zrealm_home_known_iface
package zrealm_home_known_iface

// KNOWN (T20): cbvictim's gate reaches cbvictim's own value through a interface holding a pointer
// this realm supplied; it cannot tell where the reference came from, so it writes.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunAny(cross(cur), cbvictim.Obj)
	println("owner:", cbvictim.Owner)
}

// Output:
// owner: any
