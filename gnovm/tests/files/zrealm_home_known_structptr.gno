// PKGPATH: gno.land/r/zrealm_home_known_structptr
package zrealm_home_known_structptr

// S3: the gate reaches cbvictim's own value through a struct pointer this
// realm supplied; the value was made in another crossing call, so it is read-only here.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunSt(cross(cur), cbvictim.Obj)
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>
