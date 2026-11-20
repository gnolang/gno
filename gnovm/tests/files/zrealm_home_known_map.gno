// PKGPATH: gno.land/r/zrealm_home_known_map
package zrealm_home_known_map

// S3: the gate reaches cbvictim's own value through a map this realm
// supplied; the value was made in another crossing call, so it is read-only here.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunMap(cross(cur), cbvictim.Map)
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>
