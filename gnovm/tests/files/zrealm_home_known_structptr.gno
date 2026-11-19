// PKGPATH: gno.land/r/zrealm_home_known_structptr
package zrealm_home_known_structptr

// Known gap: cbvictim's gate reaches cbvictim's own value through a struct pointer
// this realm supplied; it cannot tell where the reference came from, so it writes.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunSt(cross(cur), cbvictim.Obj)
	println("owner:", cbvictim.Owner)
}

// Output:
// owner: struct
