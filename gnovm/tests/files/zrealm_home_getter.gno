// PKGPATH: gno.land/r/zrealm_home_getter
package zrealm_home_getter

// A getter run read-only returns cbvictim's own value; this realm handing it
// to cbvictim's gate strips the maker.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.GetFV())
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): Owner<~VPBlock(3,0)>
