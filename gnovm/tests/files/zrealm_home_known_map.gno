// PKGPATH: gno.land/r/zrealm_home_known_map
package zrealm_home_known_map

// Known gap: cbvictim's gate reaches cbvictim's own value through a map
// this realm supplied; it cannot tell where the reference came from, so it writes.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunMap(cross(cur), cbvictim.Map)
	println("owner:", cbvictim.Owner)
}

// Output:
// owner: map
