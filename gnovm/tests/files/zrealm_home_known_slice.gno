// PKGPATH: gno.land/r/zrealm_home_known_slice
package zrealm_home_known_slice

// Known gap: cbvictim's gate reaches cbvictim's own value through a slice
// this realm supplied; it cannot tell where the reference came from, so it writes.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunAll(cross(cur), cbvictim.Table[0:1])
	println("owner:", cbvictim.Owner)
}

// Output:
// owner: table
